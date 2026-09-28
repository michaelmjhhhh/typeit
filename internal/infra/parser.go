package infra

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/michaelmjhhhh/typeit/internal/domain"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

//go:embed queries/*.json
var queryFiles embed.FS

type querySet struct {
	Kinds       map[string]string `json:"kinds"`
	MiddleKinds map[string]string `json:"middle_kinds"`
	Main        string            `json:"main"`
	Middle      string            `json:"middle"`
	Comments    string            `json:"comments"`
}
type capture struct {
	Node sitter.Node
	Name string
}

func captures(lang *sitter.Language, tree *sitter.Tree, source, pattern string) ([]capture, error) {
	if strings.TrimSpace(pattern) == "" {
		return nil, nil
	}
	query, qerr := sitter.NewQuery(lang, pattern)
	if qerr != nil {
		return nil, fmt.Errorf("query: %v", qerr)
	}
	defer query.Close()
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	matches := cursor.Matches(query, tree.RootNode(), []byte(source))
	var out []capture
	for match := matches.Next(); match != nil; match = matches.Next() {
		for _, c := range match.Captures {
			out = append(out, capture{Node: c.Node, Name: query.CaptureNames()[c.Index]})
		}
	}
	return out, nil
}
func Extract(path, language, source string) ([]domain.Challenge, error) {
	factory, ok := grammars[language]
	if !ok {
		return nil, fmt.Errorf("no parser for %s", language)
	}
	b, err := queryFiles.ReadFile("queries/" + language + ".json")
	if err != nil {
		return nil, err
	}
	var queries querySet
	if err = json.Unmarshal(b, &queries); err != nil {
		return nil, err
	}
	lang := sitter.NewLanguage(factory())
	parser := sitter.NewParser()
	defer parser.Close()
	if err = parser.SetLanguage(lang); err != nil {
		return nil, err
	}
	tree := parser.Parse([]byte(source), nil)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse %s", path)
	}
	defer tree.Close()
	comments, err := captures(lang, tree, source, queries.Comments)
	if err != nil {
		return nil, err
	}
	var ranges []domain.Range
	for _, c := range comments {
		ranges = append(ranges, domain.Range{utf8.RuneCountInString(source[:c.Node.StartByte()]), utf8.RuneCountInString(source[:c.Node.EndByte()])})
	}
	nodes, err := captures(lang, tree, source, queries.Main)
	if err != nil {
		return nil, err
	}
	var chunks []domain.Challenge
	for _, c := range nodes {
		kind, ok := queries.Kinds[c.Name]
		if !ok {
			continue
		}
		start, end := int(c.Node.StartByte()), int(c.Node.EndByte())
		if end-start < 10 {
			continue
		}
		chunk := nodeChunk(path, language, source, start, end, int(c.Node.StartPosition().Row)+1, int(c.Node.EndPosition().Row)+1, ranges)
		chunk.Kind = kind
		chunks = append(chunks, chunk)
		if strings.Count(chunk.Code, "\n") < 2 {
			continue
		}
		child := parser.Parse([]byte(chunk.Code), nil)
		if child == nil {
			continue
		}
		mids, e := captures(lang, child, chunk.Code, queries.Middle)
		if e != nil {
			child.Close()
			return nil, e
		}
		for _, mid := range mids {
			midKind, ok := queries.MiddleKinds[mid.Name]
			if !ok {
				continue
			}
			a, z := int(mid.Node.StartByte()), int(mid.Node.EndByte())
			content := chunk.Code[a:z]
			if z-a < 30 || z-a > 2000 || !strings.Contains(content, "\n") {
				continue
			}
			middle := nodeChunk(path, language, chunk.Code, a, z, chunk.StartLine+int(mid.Node.StartPosition().Row), chunk.StartLine+int(mid.Node.EndPosition().Row), chunk.Comments)
			middle.Kind = midKind
			chunks = append(chunks, middle)
		}
		child.Close()
	}
	sort.SliceStable(chunks, func(i, j int) bool {
		if chunks[i].StartLine == chunks[j].StartLine && chunks[i].EndLine == chunks[j].EndLine {
			return chunkPriority(chunks[i].Kind) < chunkPriority(chunks[j].Kind)
		}
		if chunks[i].StartLine == chunks[j].StartLine {
			return chunks[i].EndLine < chunks[j].EndLine
		}
		return chunks[i].StartLine < chunks[j].StartLine
	})
	var result []domain.Challenge
	seen := map[[2]int]bool{}
	for _, c := range chunks {
		key := [2]int{c.StartLine, c.EndLine}
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, domain.Generate(c, false)...)
	}
	file := domain.Challenge{Path: path, Code: source, Language: language, StartLine: 1, EndLine: len(strings.Split(strings.TrimSuffix(source, "\n"), "\n")), Comments: ranges}
	result = append(result, domain.Generate(file, true)...)
	lines := strings.Split(strings.TrimSuffix(source, "\n"), "\n")
	for i := range result {
		c := &result[i]
		a, b := max(0, c.StartLine-1), min(len(lines), c.EndLine)
		c.PreContext = lines[max(0, a-4):min(a, len(lines))]
		c.PostContext = lines[b:min(len(lines), b+4)]
	}
	return result, nil
}
func nodeChunk(path, lang, source string, start, end, startLine, endLine int, comments []domain.Range) domain.Challenge {
	lineStart := strings.LastIndex(source[:start], "\n") + 1
	prefix := source[lineStart:start]
	code := prefix + source[start:end]
	offset := utf8.RuneCountInString(source[:start])
	indent := utf8.RuneCountInString(prefix)
	endChar := utf8.RuneCountInString(source[:end])
	c := domain.Challenge{Path: path, Language: lang, Code: code, StartLine: startLine, EndLine: endLine}
	for _, r := range comments {
		a, z := max(r[0], offset), min(r[1], endChar)
		if a < z {
			c.Comments = append(c.Comments, domain.Range{a - offset + indent, z - offset + indent})
		}
	}
	return c
}

func chunkPriority(kind string) int {
	switch kind {
	case "Function", "Class", "Method":
		return 0
	case "CodeBlock":
		return 10
	default:
		return 5
	}
}
