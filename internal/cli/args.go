package cli

import (
	"fmt"
	"strings"
)

type Args struct {
	Path, Repo, Command, Action, Format, Output, Language, RepoName, Period string
	Languages                                                               []string
	Force, Help, Version                                                    bool
}

const Help = `GitType — typing practice with your own source code

Usage: gittype [OPTIONS] [REPO_PATH] [COMMAND]

Options:
  --repo <REPOSITORY>   GitHub owner/repo, HTTPS URL, or SSH URL
  --langs <LANGUAGES>   Comma-separated language filter
  -h, --help           Show help
  -V, --version        Show version

Commands:
  history                         Session history
  stats                           Analytics
  export [--format json|csv] [--output FILE]
  cache stats|list|clear           Manage extracted challenges
  repo list|play|clear [--force]    Manage cached repositories
  trending [LANGUAGE] [REPO] [--period daily|weekly|monthly]

Examples:
  gittype . --langs go,rust
  gittype --repo unhappychoice/gittype
  gittype trending rust --period weekly

Data is stored in ~/.gittype (override with GITTYPE_DATA_DIR).
`

func Parse(arguments []string) (Args, error) {
	a := Args{Path: ".", Format: "json", Period: "daily"}
	var positional []string
	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]
		if arg == "--" {
			positional = append(positional, arguments[i+1:]...)
			break
		}
		name, value, has := strings.Cut(arg, "=")
		switch name {
		case "--help", "-h":
			a.Help = true
		case "--version", "-V":
			a.Version = true
		case "--force":
			a.Force = true
		case "--repo", "--langs", "--format", "--output", "--period":
			if !has {
				i++
				if i >= len(arguments) {
					return a, fmt.Errorf("%s requires a value", name)
				}
				value = arguments[i]
			}
			if value == "" {
				return a, fmt.Errorf("%s requires a nonempty value", name)
			}
			switch name {
			case "--repo":
				a.Repo = value
			case "--langs":
				a.Languages = strings.Split(value, ",")
			case "--format":
				a.Format = value
			case "--output":
				a.Output = value
			case "--period":
				a.Period = value
			}
		default:
			if strings.HasPrefix(arg, "-") {
				return a, fmt.Errorf("unknown option %s", arg)
			}
			positional = append(positional, arg)
		}
	}
	commands := map[string]bool{"history": true, "stats": true, "export": true, "cache": true, "repo": true, "trending": true}
	if len(positional) > 0 && !commands[positional[0]] {
		a.Path = positional[0]
		positional = positional[1:]
	}
	if len(positional) > 0 {
		a.Command = positional[0]
		positional = positional[1:]
		if !commands[a.Command] {
			return a, fmt.Errorf("unexpected argument %s", a.Command)
		}
	}
	switch a.Command {
	case "cache", "repo":
		if len(positional) == 0 {
			if a.Help {
				return a, nil
			}
			return a, fmt.Errorf("%s requires a subcommand", a.Command)
		}
		a.Action = positional[0]
		positional = positional[1:]
		valid := a.Action == "list" || a.Action == "clear" || (a.Command == "cache" && a.Action == "stats") || (a.Command == "repo" && a.Action == "play")
		if !valid {
			return a, fmt.Errorf("unknown %s subcommand %s", a.Command, a.Action)
		}
	case "trending":
		if len(positional) > 0 {
			a.Language = positional[0]
			positional = positional[1:]
		}
		if len(positional) > 0 {
			a.RepoName = positional[0]
			positional = positional[1:]
		}
	}
	if len(positional) > 0 {
		return a, fmt.Errorf("unexpected argument %s", positional[0])
	}
	if a.Format != "json" && a.Format != "csv" {
		return a, fmt.Errorf("unsupported export format %s", a.Format)
	}
	if a.Period != "daily" && a.Period != "weekly" && a.Period != "monthly" {
		return a, fmt.Errorf("invalid period %s", a.Period)
	}
	if a.Force && (a.Command != "repo" || a.Action != "clear") {
		return a, fmt.Errorf("--force is only valid for repo clear")
	}
	return a, nil
}
