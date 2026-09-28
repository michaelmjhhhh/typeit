package domain

import (
	"strings"
	"unicode"
)

type Range [2]int

type Challenge struct {
	PreContext  []string   `json:"pre_context,omitempty"`
	PostContext []string   `json:"post_context,omitempty"`
	Kind        string     `json:"-"`
	ID          string     `json:"id"`
	Path        string     `json:"source_file_path"`
	Code        string     `json:"code_content"`
	StartLine   int        `json:"start_line"`
	EndLine     int        `json:"end_line"`
	Language    string     `json:"language"`
	Comments    []Range    `json:"comment_ranges"`
	Difficulty  Difficulty `json:"difficulty_level"`
}

type Typing struct {
	Display    []rune
	DisplayMap []int
	Ranges     []Range
	Original   []rune
	Text       []rune
	Mapping    []int
	Position   int
	Mistakes   int
	Wrong      bool
}

func InComment(pos int, ranges []Range) bool {
	for _, r := range ranges {
		if pos >= r[0] && pos < r[1] {
			return true
		}
	}
	return false
}

func NewTyping(c Challenge) *Typing {
	c.Comments = NormalizeRanges(c.Code, c.Comments)
	t := &Typing{Original: []rune(c.Code), Ranges: c.Comments}
	previousEnd := 0
	offset := 0
	for _, line := range strings.Split(c.Code, "\n") {
		chars := []rune(line)
		var positions []int
		for i, r := range chars {
			if !InComment(offset+i, c.Comments) && r != '\r' {
				positions = append(positions, offset+i)
			}
		}
		for len(positions) > 0 && unicode.IsSpace(t.Original[positions[0]]) {
			positions = positions[1:]
		}
		for len(positions) > 0 && unicode.IsSpace(t.Original[positions[len(positions)-1]]) {
			positions = positions[:len(positions)-1]
		}
		if len(positions) > 0 {
			if len(t.Text) > 0 {
				t.Text = append(t.Text, '\n')
				t.Mapping = append(t.Mapping, previousEnd)
			}
			for _, p := range positions {
				t.Text = append(t.Text, t.Original[p])
				t.Mapping = append(t.Mapping, p)
			}
		}
		if len(positions) > 0 {
			previousEnd = offset + len(chars)
		}
		offset += len(chars) + 1
	}
	t.buildDisplay()
	return t
}

func (t *Typing) Completed() bool { return t.Position >= len(t.Text) }
func (t *Typing) Input(r rune) bool {
	if t.Completed() {
		return false
	}
	if r == t.Text[t.Position] {
		t.Position++
		t.Wrong = false
		return true
	}
	t.Mistakes++
	t.Wrong = true
	return false
}
func (t *Typing) DisplayPosition() int {
	if t.Completed() {
		return len(t.Original)
	}
	return t.Mapping[t.Position]
}
