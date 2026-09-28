package cli

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/michaelmjhhhh/typeit/internal/infra"
	"github.com/michaelmjhhhh/typeit/internal/ui"
)

var Version = "0.10.2-go"

func Run(ctx context.Context, arguments []string, in io.Reader, out io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	a, err := Parse(arguments)
	if err != nil {
		return err
	}
	if a.Help {
		_, err = fmt.Fprint(out, Help)
		return err
	}
	if a.Version {
		_, err = fmt.Fprintln(out, "gittype "+Version)
		return err
	}
	for _, name := range a.Languages {
		if _, err = infra.ResolveLanguage(name); err != nil {
			return err
		}
	}
	dir, err := infra.DataDir()
	if err != nil {
		return err
	}
	if a.Command == "cache" {
		return cacheCommand(dir, a.Action, out)
	}
	if a.Command == "repo" && a.Action != "play" && (a.Action != "list" || !term.IsTerminal(os.Stdout.Fd())) {
		return repoCommand(dir, a, in, out)
	}
	db, err := infra.OpenDatabase(dir)
	if err != nil {
		return err
	}
	defer db.Close()
	if a.Command == "export" {
		return exportCommand(db, a, out)
	}
	if !term.IsTerminal(os.Stdout.Fd()) || !term.IsTerminal(os.Stdin.Fd()) {
		return fmt.Errorf("interactive mode requires a terminal; use --help, export, cache or repo list")
	}
	config, err := infra.LoadConfig(dir)
	if err != nil {
		return err
	}
	themes, err := infra.Themes(dir)
	if err != nil {
		return err
	}
	screen := a.Command
	if a.Command == "repo" {
		screen = "repos"
		if a.Action == "list" {
			screen = "repo-list"
		}
	}
	options := ui.Options{Version: strings.TrimSuffix(Version, "-go"), Context: ctx, DataDir: dir, Path: a.Path, Remote: a.Repo, Languages: a.Languages, Screen: screen, Language: a.Language, Period: a.Period, RepoName: a.RepoName, Database: db, Config: config, Themes: themes}
	program := tea.NewProgram(ui.New(options), tea.WithAltScreen(), tea.WithContext(ctx), tea.WithInput(in), tea.WithOutput(out))
	_, err = program.Run()
	return err
}
func cacheCommand(dir, action string, out io.Writer) error {
	if action == "clear" {
		if err := os.RemoveAll(filepath.Join(dir, "challenge-cache")); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "Challenge cache cleared.")
		return err
	}
	entries, err := infra.CacheStats(dir)
	if err != nil {
		return err
	}
	var size int64
	count := 0
	for _, e := range entries {
		size += e.Size
		count += e.Count
		if action == "list" {
			fmt.Fprintf(out, "%s  %s  %d challenges\n", e.Key, e.Path, e.Count)
		}
	}
	_, err = fmt.Fprintf(out, "%d cached repositories · %d challenges · %d bytes\n", len(entries), count, size)
	return err
}
func repoCommand(dir string, a Args, in io.Reader, out io.Writer) error {
	repos, err := infra.CachedRepositories(dir)
	if err != nil {
		return err
	}
	if a.Action == "list" {
		for _, r := range repos {
			fmt.Fprintf(out, "%s/%s  %s\n", r.Owner, r.Name, r.Path)
		}
		_, err = fmt.Fprintf(out, "%d cached repositories\n", len(repos))
		return err
	}
	if !a.Force {
		fmt.Fprintf(out, "Clear all %d cached repositories? [y/N] ", len(repos))
		line, err := bufio.NewReader(in).ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if strings.ToLower(strings.TrimSpace(line)) != "y" && strings.ToLower(strings.TrimSpace(line)) != "yes" {
			_, err = fmt.Fprintln(out, "Cancelled.")
			return err
		}
	}
	if err = os.RemoveAll(filepath.Join(dir, "repos")); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "Repository cache cleared.")
	return err
}
func exportCommand(db *infra.Database, a Args, out io.Writer) error {
	sessions, err := db.History()
	if err != nil {
		return err
	}
	for i := range sessions {
		sessions[i].Stages, err = db.Stages(sessions[i].ID)
		if err != nil {
			return err
		}
	}
	var buffer strings.Builder
	if a.Format == "json" {
		encoder := json.NewEncoder(&buffer)
		encoder.SetIndent("", "  ")
		if err = encoder.Encode(sessions); err != nil {
			return err
		}
	} else {
		writer := csv.NewWriter(&buffer)
		if err = writer.Write([]string{"id", "started_at", "repository", "difficulty", "wpm", "cpm", "accuracy", "score", "mistakes", "duration_ms", "stages_completed"}); err != nil {
			return err
		}
		for _, s := range sessions {
			row := []string{strconv.FormatInt(s.ID, 10), s.StartedAt, s.Repository.Owner + "/" + s.Repository.Name, string(s.Difficulty), fmt.Sprint(s.WPM), fmt.Sprint(s.CPM), fmt.Sprint(s.Accuracy), fmt.Sprint(s.Score), fmt.Sprint(s.Mistakes), fmt.Sprint(s.DurationMS), fmt.Sprint(s.Completed)}
			if err = writer.Write(row); err != nil {
				return err
			}
		}
		writer.Flush()
		if err = writer.Error(); err != nil {
			return err
		}
	}
	if a.Output != "" {
		return infra.AtomicWrite(a.Output, []byte(buffer.String()))
	}
	_, err = io.WriteString(out, buffer.String())
	return err
}
