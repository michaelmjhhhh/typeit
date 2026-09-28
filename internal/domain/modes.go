package domain

import (
	"math/rand/v2"
	"sort"
	"strings"
	"time"
)

type GameMode string

const (
	NormalMode GameMode = "Normal"
	TimeAttack GameMode = "TimeAttack"
	CustomMode GameMode = "Custom"
)

type StageConfig struct {
	Mode       GameMode
	MaxStages  int
	Difficulty Difficulty
	TimeLimit  time.Duration
	Seed       uint64
}

func BuildStages(challenges []Challenge, config StageConfig) []Challenge {
	out := append([]Challenge{}, challenges...)
	count := config.MaxStages
	if count <= 0 {
		count = 3
	}
	if config.Mode == CustomMode {
		out = nil
		for _, c := range challenges {
			if c.Difficulty == config.Difficulty {
				out = append(out, c)
			}
		}
	}
	if config.Mode == TimeAttack {
		sort.SliceStable(out, func(i, j int) bool { return lineCount(out[i].Code) < lineCount(out[j].Code) })
		return out
	}
	random := rand.New(rand.NewPCG(config.Seed, config.Seed+1))
	random.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	if config.Mode != CustomMode {
		penalty := func(c Challenge) int {
			n := lineCount(c.Code)
			if n < 5 {
				return n + 100
			}
			if n > 20 {
				return n + 50
			}
			return n
		}
		sort.SliceStable(out, func(i, j int) bool { return penalty(out[i]) < penalty(out[j]) })
	}
	return out[:min(count, len(out))]
}
func lineCount(code string) int {
	if code == "" {
		return 0
	}
	return len(strings.Split(strings.TrimSuffix(code, "\n"), "\n"))
}
