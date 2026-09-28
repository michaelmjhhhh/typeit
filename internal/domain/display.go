package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func NormalizeRanges(code string, ranges []Range) []Range {
	chars := []rune(code)
	out := make([]Range, 0, len(ranges))
	looks := func(s string) bool {
		return strings.HasPrefix(s, "//") || strings.HasPrefix(s, "/*") || strings.HasPrefix(s, "#")
	}
	for _, r := range ranges {
		a, b := r[0], r[1]
		if a < 0 || b <= a {
			continue
		}
		byteLike := a <= len(code) && b <= len(code) && utf8.ValidString(code[:a]) && utf8.ValidString(code[:b]) && looks(code[a:b])
		charLike := b <= len(chars) && looks(string(chars[a:b]))
		if (byteLike && !charLike) || b > len(chars) {
			a = utf8.RuneCountInString(code[:min(a, len(code))])
			b = utf8.RuneCountInString(code[:min(b, len(code))])
		}
		if a < len(chars) {
			out = append(out, Range{a, min(b, len(chars))})
		}
	}
	return out
}
func (t *Typing) buildDisplay() {
	offset := 0
	lines := strings.Split(strings.TrimSuffix(string(t.Original), "\n"), "\n")
	for lineIndex, line := range lines {
		chars := []rune(strings.TrimSuffix(line, "\r"))
		last := -1
		for i, r := range chars {
			if !unicode.IsSpace(r) && !InComment(offset+i, t.Ranges) {
				last = i
			}
		}
		for i, r := range chars {
			display := []rune{r}
			if r == '\t' {
				display = []rune("→   ")
			}
			for _, v := range display {
				t.Display = append(t.Display, v)
				t.DisplayMap = append(t.DisplayMap, offset+i)
			}
			if i == last {
				t.Display = append(t.Display, '↵')
				t.DisplayMap = append(t.DisplayMap, offset+len(chars))
			}
		}
		if lineIndex < len(lines)-1 {
			t.Display = append(t.Display, '\n')
			t.DisplayMap = append(t.DisplayMap, offset+len(chars))
		}
		offset += len([]rune(line)) + 1
	}
}
func (t *Typing) DisplayCursor() int {
	target := t.DisplayPosition()
	for i, p := range t.DisplayMap {
		if p >= target {
			return i
		}
	}
	return len(t.Display)
}
