package gittype_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelmjhhhh/typeit/internal/infra"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestTrendingCacheAndVersion(t *testing.T) {
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	requests := 0
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		body := ""
		switch r.URL.Host {
		case "api.ossinsight.io":
			if r.URL.Query().Get("language") != "Rust" || r.URL.Query().Get("period") != "past_week" {
				t.Errorf("bad trending URL %s", r.URL)
			}
			body = `{"data":{"rows":[{"repo_name":"owner/repo","primary_language":"Rust","description":"sample","stars":"42"}]}}`
		case "api.github.com":
			body = `{"tag_name":"v0.10.3"}`
		default:
			t.Fatalf("unexpected request %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	dir := t.TempDir()
	for i := 0; i < 2; i++ {
		repos, err := infra.Trending(context.Background(), dir, "rust", "weekly")
		if err != nil || len(repos) != 1 || repos[0].Name != "owner/repo" {
			t.Fatalf("%v %v", repos, err)
		}
	}
	if requests != 1 {
		t.Fatal("trending cache missed")
	}
	for i := 0; i < 2; i++ {
		v, err := infra.CheckVersion(context.Background(), dir, "0.10.2")
		if err != nil || !v.Available || v.Latest != "0.10.3" {
			t.Fatalf("%+v %v", v, err)
		}
	}
	if requests != 2 {
		t.Fatal("version cache missed")
	}
}
func TestCloneAndReuseRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	root := t.TempDir()
	source := filepath.Join(root, "owner", "sample.git")
	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	git := func(input string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", source}, args...)...)
		cmd.Stdin = strings.NewReader(input)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=GitType Test", "GIT_AUTHOR_EMAIL=test@example.invalid", "GIT_COMMITTER_NAME=GitType Test", "GIT_COMMITTER_EMAIL=test@example.invalid")
		b, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, b)
		}
		return strings.TrimSpace(string(b))
	}
	git("", "init", "--bare", "--initial-branch=main")
	blob := git(languageSamples["go"], "hash-object", "-w", "--stdin")
	tree := git(fmt.Sprintf("100644 blob %s\tmain.go\n", blob), "mktree")
	commit := git("fixture\n", "commit-tree", tree)
	git("", "update-ref", "refs/heads/main", commit)
	git("", "update-server-info")
	server := httptest.NewServer(http.FileServer(http.Dir(root)))
	defer server.Close()
	cache := t.TempDir()
	path, err := infra.Clone(context.Background(), cache, server.URL+"/owner/sample.git")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(path, "main.go"))
	if err != nil || string(b) != languageSamples["go"] {
		t.Fatal("clone source mismatch", err)
	}
	repos, err := infra.CachedRepositories(cache)
	if err != nil || len(repos) != 1 || repos[0].Path != path {
		t.Fatalf("cached repos %v %v", repos, err)
	}
	server.Close()
	again, err := infra.Clone(context.Background(), cache, server.URL+"/owner/sample.git")
	if err != nil || again != path {
		t.Fatal("cache not reused", err)
	}
}
func TestRepositoryFormats(t *testing.T) {
	for _, input := range []string{"https://gitlab.com/owner/repo.git", "ssh://git@example.com:2222/owner/repo.git", "git@example.com:2222:owner/repo.git", "https://gitlab.com/group/subgroup/repo"} {
		r, err := infra.ParseRepoRef(input)
		if err != nil || r.Owner == "" || r.Name == "" {
			t.Fatalf("%s: %+v %v", input, r, err)
		}
	}
}
