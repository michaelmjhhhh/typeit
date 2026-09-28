package gittype_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/michaelmjhhhh/typeit/internal/cli"
)

func TestCLIArguments(t *testing.T) {
	cases := []struct {
		args            []string
		command, action string
	}{{[]string{"--langs", "go,rust", "."}, "", ""}, {[]string{"repo", "clear", "--force"}, "repo", "clear"}, {[]string{"trending", "rust", "owner/repo", "--period=weekly"}, "trending", ""}, {[]string{"export", "--format", "csv"}, "export", ""}}
	for _, c := range cases {
		a, err := cli.Parse(c.args)
		if err != nil || a.Command != c.command || a.Action != c.action {
			t.Fatalf("%v: %+v %v", c.args, a, err)
		}
	}
	for _, args := range [][]string{{"--repo"}, {"--unknown"}, {"cache"}, {"repo", "bad"}, {"trending", "--period", "yearly"}, {"export", "--format", "xml"}} {
		if _, err := cli.Parse(args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
func TestNoninteractiveCLI(t *testing.T) {
	t.Setenv("TYPEIT_DATA_DIR", t.TempDir())
	for _, args := range [][]string{{"--help"}, {"--version"}, {"cache", "stats"}, {"cache", "list"}, {"repo", "list"}, {"repo", "clear"}, {"export"}} {
		var out bytes.Buffer
		if err := cli.Run(context.Background(), args, strings.NewReader("n\n"), &out); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if out.Len() == 0 {
			t.Fatalf("no output for %v", args)
		}
		if args[0] == "export" && !json.Valid(out.Bytes()) {
			t.Fatalf("invalid export %q", out.String())
		}
	}
}
