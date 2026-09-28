package domain

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode"
)

type Difficulty string

const (
	Easy   Difficulty = "Easy"
	Normal Difficulty = "Normal"
	Hard   Difficulty = "Hard"
	Wild   Difficulty = "Wild"
	Zen    Difficulty = "Zen"
)

var Difficulties = []Difficulty{Easy, Normal, Hard, Wild, Zen}

func (d Difficulty) Limits() (int, int) {
	switch d {
	case Easy:
		return 20, 100
	case Normal:
		return 80, 200
	case Hard:
		return 180, 500
	default:
		return 0, int(^uint(0) >> 1)
	}
}
func CodeCharacters(c Challenge) int {
	n := 0
	for i, r := range []rune(c.Code) {
		if !unicode.IsSpace(r) && !InComment(i, c.Comments) {
			n++
		}
	}
	return n
}
func Generate(c Challenge, file bool) []Challenge {
	var result []Challenge
	for _, d := range Difficulties {
		if d == Zen && !file {
			continue
		}
		v := c
		v.Difficulty = d
		minChars, maxChars := d.Limits()
		count := CodeCharacters(v)
		if count < minChars {
			continue
		}
		if count > maxChars {
			lines := strings.Split(strings.TrimSuffix(v.Code, "\n"), "\n")
			offset, total, last, cut := 0, 0, 0, len(lines)
			for i, line := range lines {
				for j, r := range []rune(line) {
					if !unicode.IsSpace(r) && !InComment(offset+j, v.Comments) {
						total++
					}
				}
				if total > maxChars {
					cut = max(1, last)
					break
				}
				s := strings.TrimSpace(line)
				if s == "" || strings.ContainsAny(stringLast(s), "}]);") {
					last = i + 1
				}
				offset += len([]rune(line)) + 1
			}
			v.Code = strings.TrimRightFunc(strings.Join(lines[:cut], "\n"), unicode.IsSpace)
			v.EndLine = v.StartLine + cut - 1
			v.Comments = nil
			n := len([]rune(v.Code))
			for _, r := range c.Comments {
				if r[0] < n {
					v.Comments = append(v.Comments, Range{r[0], min(n, r[1])})
				}
			}
			if CodeCharacters(v) < minChars {
				continue
			}
		}
		if len(NewTyping(v).Text) == 0 {
			continue
		}
		v.ID = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s:%s", v.Path, v.StartLine, d, v.Code))))
		result = append(result, v)
	}
	return result
}
func stringLast(s string) string {
	if s == "" {
		return ""
	}
	return s[len(s)-1:]
}
