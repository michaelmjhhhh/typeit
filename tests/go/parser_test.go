package gittype_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelmjhhhh/typeit/internal/domain"
	"github.com/michaelmjhhhh/typeit/internal/infra"
)

var languageSamples = map[string]string{
	"c":          "// comment\nint add(int a, int b) {\n return a + b;\n}\n",
	"cpp":        "// comment\nclass Counter {\n public: int value() { return 42; }\n};\n",
	"csharp":     "// comment\nclass Counter {\n public int Value() { return 42; }\n}\n",
	"clojure":    "; comment\n(defn greeting [name]\n  (str \"hello \" name))\n",
	"dart":       "// comment\nint add(int a, int b) {\n return a + b;\n}\n",
	"elixir":     "# comment\ndefmodule Counter do\n def add(a, b) do\n  a + b\n end\nend\n",
	"erlang":     "% comment\n-module(counter).\n-export([add/2]).\nadd(A, B) ->\n A + B.\n",
	"go":         "package main\n// comment\nfunc add(a, b int) int {\n return a + b\n}\n",
	"haskell":    "-- comment\nmodule Main where\nadd a b = a + b\nmain = print (add 1 2)\n",
	"java":       "// comment\nclass Counter {\n public int add(int a, int b) { return a + b; }\n}\n",
	"javascript": "// comment\nfunction add(a, b) {\n return a + b;\n}\n",
	"kotlin":     "// comment\nfun add(a: Int, b: Int): Int {\n return a + b\n}\n",
	"php":        "<?php\n// comment\nfunction add($a, $b) {\n return $a + $b;\n}\n",
	"python":     "# comment\ndef add(a, b):\n    return a + b\n",
	"ruby":       "# comment\ndef add(a, b)\n  a + b\nend\n",
	"rust":       "// comment\nfn add(a: i32, b: i32) -> i32 {\n a + b\n}\n",
	"scala":      "// comment\nobject Counter {\n def add(a: Int, b: Int): Int = { a + b }\n}\n",
	"swift":      "// comment\nfunc add(_ a: Int, _ b: Int) -> Int {\n return a + b\n}\n",
	"typescript": "// comment\nfunction add(a: number, b: number): number {\n return a + b;\n}\n",
	"zig":        "// comment\nfn add(a: i32, b: i32) i32 {\n return a + b;\n}\n",
}

func TestAllLanguageGrammars(t *testing.T) {
	if len(infra.Languages()) != 20 {
		t.Fatal("missing languages")
	}
	for lang, source := range languageSamples {
		t.Run(lang, func(t *testing.T) {
			challenges, err := infra.Extract("sample", lang, source)
			if err != nil {
				t.Fatal(err)
			}
			zen, wild := false, false
			for _, c := range challenges {
				if c.Difficulty == domain.Zen {
					zen = true
					if strings.Contains(string(domain.NewTyping(c).Text), "comment") {
						t.Fatal("comment not removed")
					}
				}
				if c.Difficulty == domain.Wild && c.Code != source {
					wild = true
				}
			}
			if !zen || !wild {
				t.Fatalf("zen=%v extracted chunk=%v (%d challenges)", zen, wild, len(challenges))
			}
		})
	}
}
func TestScanIgnoreAndUnicode(t *testing.T) {
	dir := t.TempDir()
	write := func(path, content string) {
		t.Helper()
		p := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("main.go", languageSamples["go"])
	write("ignored/hidden.go", languageSamples["go"])
	write("vendor/hidden.go", languageSamples["go"])
	write("other.py", languageSamples["python"])
	write(".gittypeignore", "/ignored/\n/vendor/\n")
	challenges, err := infra.Scan(context.Background(), dir, []string{"Go"})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range challenges {
		if c.Path != "main.go" {
			t.Fatalf("unexpected path %q", c.Path)
		}
	}
	if _, err = infra.Scan(context.Background(), dir, []string{"not-a-language"}); err == nil {
		t.Fatal("accepted bad language")
	}
}
func TestCacheInvalidatesChanges(t *testing.T) {
	root, cache := t.TempDir(), t.TempDir()
	file := filepath.Join(root, "main.go")
	if err := os.WriteFile(file, []byte(languageSamples["go"]), 0644); err != nil {
		t.Fatal(err)
	}
	first, err := infra.LoadChallenges(context.Background(), cache, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(languageSamples["go"], "add", "multiply")
	if err = os.WriteFile(file, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	second, err := infra.LoadChallenges(context.Background(), cache, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first[0].ID == second[0].ID {
		t.Fatal("stale cache")
	}
}
func TestRepositoryParsing(t *testing.T) {
	for _, ref := range []string{"owner/repo", "https://github.com/owner/repo.git", "git@github.com:owner/repo.git"} {
		parsed, err := infra.ParseRepoRef(ref)
		if err != nil || parsed.Owner != "owner" || parsed.Name != "repo" {
			t.Fatalf("%s: %s %s %v", ref, parsed.Owner, parsed.Name, err)
		}
	}
	for _, ref := range []string{"../repo", "owner/..", "-flag/repo"} {
		if _, err := infra.ParseRepoRef(ref); err == nil {
			t.Fatalf("accepted %s", ref)
		}
	}
}

func TestInheritedIgnoreRulesAndCache(t *testing.T) {
	root, cache := t.TempDir(), t.TempDir()
	write := func(path, text string) {
		t.Helper()
		p := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	git := func(args ...string) {
		t.Helper()
		b, e := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
		if e != nil {
			t.Fatalf("%s: %v", b, e)
		}
	}
	git("init", "--quiet")
	for _, name := range []string{"keep", "global", "info", "ancestor", "local", "cache"} {
		write("sub/"+name+".go", languageSamples["go"])
	}
	write("global.ignore", "global.go\n")
	git("config", "core.excludesfile", filepath.Join(root, "global.ignore"))
	write(".git/info/exclude", "info.go\n")
	write(".gitignore", "/sub/ancestor.go\n")
	write("sub/.ignore", "\ufefflocal.go\n")
	scan := func() map[string]bool {
		t.Helper()
		cs, e := infra.LoadChallenges(context.Background(), cache, filepath.Join(root, "sub"), nil)
		if e != nil {
			t.Fatal(e)
		}
		paths := map[string]bool{}
		for _, c := range cs {
			paths[c.Path] = true
		}
		return paths
	}
	paths := scan()
	if len(paths) != 2 || !paths["keep.go"] || !paths["cache.go"] {
		t.Fatalf("ignore rules: %v", paths)
	}
	write(".git/info/exclude", "info.go\ncache.go\n")
	paths = scan()
	if len(paths) != 1 || !paths["keep.go"] {
		t.Fatalf("exclude change did not invalidate cache: %v", paths)
	}
}
func TestFourLinesOfSourceContext(t *testing.T) {
	source := "package main\n// one\n// two\n// three\n// four\nfunc add(a, b int) int { return a+b }\n// five\n// six\n// seven\n// eight\n"
	cs, err := infra.Extract("main.go", "go", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cs {
		if c.Difficulty == domain.Wild {
			if len(c.PreContext) != 4 || len(c.PostContext) != 4 {
				t.Fatalf("context %v %v", c.PreContext, c.PostContext)
			}
			return
		}
	}
	t.Fatal("missing function challenge")
}
