package infra

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type RepoRef struct{ Origin, Owner, Name, URL string }

var repoPart = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var hostPart = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.:-]*$`)

func ParseRepo(ref string) (string, string, error) {
	r, err := ParseRepoRef(ref)
	return r.Owner, r.Name, err
}
func ParseRepoRef(input string) (RepoRef, error) {
	ref := strings.TrimSpace(input)
	r := RepoRef{Origin: "github.com"}
	path := ref
	switch {
	case strings.Contains(ref, "://"):
		u, err := url.Parse(ref)
		if err != nil {
			return r, err
		}
		if u.Scheme != "https" && u.Scheme != "http" && u.Scheme != "ssh" {
			return r, fmt.Errorf("unsupported repository URL scheme")
		}
		if u.RawQuery != "" || u.Fragment != "" {
			return r, fmt.Errorf("repository URL cannot have query or fragment")
		}
		r.Origin = u.Host
		path = strings.Trim(u.Path, "/")
		r.URL = ref
	case strings.Contains(ref, "@"):
		_, after, ok := strings.Cut(ref, "@")
		if !ok {
			return r, fmt.Errorf("invalid SSH repository")
		}
		host, rest, ok := strings.Cut(after, ":")
		if !ok {
			return r, fmt.Errorf("invalid SSH repository")
		}
		r.Origin = host
		path = rest
		if port, tail, has := strings.Cut(rest, ":"); has {
			if _, e := strconv.ParseUint(port, 10, 16); e == nil {
				path = tail
				r.Origin = host + ":" + port
				r.URL = "ssh://git@" + r.Origin + "/" + path
			}
		}
		if r.URL == "" {
			r.URL = ref
		}
	}
	path = strings.TrimSuffix(strings.TrimSuffix(path, "/"), ".git")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || !hostPart.MatchString(r.Origin) {
		return r, fmt.Errorf("invalid repository %q; use owner/repo", input)
	}
	for _, p := range parts {
		if !repoPart.MatchString(p) || p == "." || p == ".." || strings.HasPrefix(p, "-") {
			return r, fmt.Errorf("invalid repository path %q", path)
		}
	}
	r.Owner = parts[0]
	r.Name = strings.Join(parts[1:], "/")
	if r.URL == "" {
		r.URL = "https://" + r.Origin + "/" + path + ".git"
	}
	return r, nil
}
