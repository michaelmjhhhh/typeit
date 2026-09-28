package infra

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/michaelmjhhhh/typeit/internal/domain"
)

//go:embed schema.sql
var schema string

type Database struct{ DB *sql.DB }

func OpenDatabase(dir string) (*Database, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "typeit.db")
	if err := copyLegacyDatabase(filepath.Join(dir, "gittype.db"), path); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", sqliteURI(path)+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec(schema + "\nCREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY, applied_at DATETIME DEFAULT CURRENT_TIMESTAMP); INSERT OR IGNORE INTO schema_version(version) VALUES(1);"); err != nil {
		db.Close()
		return nil, err
	}
	return &Database{DB: db}, nil
}
func (db *Database) Close() error { return db.DB.Close() }
func (db *Database) Save(s *domain.SessionResult) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	repo := s.Repository
	if _, err = tx.Exec("INSERT OR IGNORE INTO repositories(user_name,repository_name,remote_url) VALUES(?,?,?)", repo.Owner, repo.Name, repo.URL); err != nil {
		return err
	}
	var rid int64
	if err = tx.QueryRow("SELECT id FROM repositories WHERE user_name=? AND repository_name=?", repo.Owner, repo.Name).Scan(&rid); err != nil {
		return err
	}
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	res, err := tx.Exec("INSERT INTO sessions(repository_id,started_at,completed_at,branch,commit_hash,is_dirty,game_mode,difficulty_level,max_stages) VALUES(?,?,?,?,?,?,?,?,?)", rid, sqliteTimestamp(s.StartedAt), now, repo.Branch, repo.Commit, repo.Dirty, s.Mode, s.Difficulty, 3)
	if err != nil {
		return err
	}
	sid, err := res.LastInsertId()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO session_results(session_id,repository_id,keystrokes,mistakes,duration_ms,wpm,cpm,accuracy,stages_completed,stages_attempted,stages_skipped,partial_effort_keystrokes,partial_effort_mistakes,score,rank_name,tier_name,game_mode,difficulty_level) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, sid, rid, s.Keystrokes, s.Mistakes, s.DurationMS, s.WPM, s.CPM, s.Accuracy, s.Completed, s.Attempted, s.Skipped, s.PartialKeys, s.PartialMistakes, s.Score, s.RankName, s.TierName, s.Mode, s.Difficulty)
	if err != nil {
		return err
	}
	if err = saveSessionMetadata(tx, sid, *s); err != nil {
		return err
	}
	for i, stage := range s.Stages {
		if err = saveStage(tx, sid, rid, i+1, stage, now); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	s.ID = sid
	return nil
}
func saveStage(tx *sql.Tx, sid, rid int64, number int, s domain.StageResult, now string) error {
	c := s.Challenge
	ranges, err := json.Marshal(c.Comments)
	if err != nil {
		return err
	}
	streaks, err := json.Marshal(s.Streaks)
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT OR IGNORE INTO challenges(id,file_path,start_line,end_line,language,code_content,comment_ranges,difficulty_level) VALUES(?,?,?,?,?,?,?,?)", c.ID, c.Path, c.StartLine, c.EndLine, c.Language, c.Code, string(ranges), c.Difficulty)
	if err != nil {
		return err
	}
	res, err := tx.Exec("INSERT INTO stages(session_id,challenge_id,stage_number,completed_at) VALUES(?,?,?,?)", sid, c.ID, number, now)
	if err != nil {
		return err
	}
	stageID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO stage_results(stage_id,session_id,repository_id,keystrokes,mistakes,duration_ms,wpm,cpm,accuracy,consistency_streaks,score,rank_name,tier_name,was_skipped,was_failed,completed_at,language,difficulty_level) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, stageID, sid, rid, s.Keystrokes, s.Mistakes, s.DurationMS, s.WPM, s.CPM, s.Accuracy, string(streaks), s.Score, s.RankName, s.TierName, s.Skipped, s.Failed, now, c.Language, c.Difficulty)
	if err != nil {
		return err
	}
	tp, tt, p, total := domain.RankPositions(s.Score)
	_, err = tx.Exec("UPDATE stage_results SET rank_position=?,rank_total=?,position=?,total=? WHERE stage_id=?", tp, tt, p, total, stageID)
	return err
}
func (db *Database) History() ([]domain.SessionResult, error) {
	rows, err := db.DB.Query(`SELECT s.id,s.started_at,COALESCE(s.game_mode,''),COALESCE(s.difficulty_level,''),COALESCE(r.user_name,''),COALESCE(r.repository_name,''),COALESCE(r.remote_url,''),COALESCE(s.branch,''),COALESCE(s.commit_hash,''),COALESCE(s.is_dirty,0),sr.keystrokes,sr.mistakes,sr.duration_ms,COALESCE(sr.wpm,0),COALESCE(sr.cpm,0),COALESCE(sr.accuracy,0),COALESCE(sr.score,0),COALESCE(sr.rank_name,''),COALESCE(sr.tier_name,''),sr.stages_completed,sr.stages_attempted,sr.stages_skipped,COALESCE(sr.partial_effort_keystrokes,0),COALESCE(sr.partial_effort_mistakes,0) FROM sessions s JOIN session_results sr ON sr.session_id=s.id LEFT JOIN repositories r ON r.id=s.repository_id ORDER BY s.started_at DESC,s.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.SessionResult{}
	for rows.Next() {
		var s domain.SessionResult
		err = rows.Scan(&s.ID, &s.StartedAt, &s.Mode, &s.Difficulty, &s.Repository.Owner, &s.Repository.Name, &s.Repository.URL, &s.Repository.Branch, &s.Repository.Commit, &s.Repository.Dirty, &s.Keystrokes, &s.Mistakes, &s.DurationMS, &s.WPM, &s.CPM, &s.Accuracy, &s.Score, &s.RankName, &s.TierName, &s.Completed, &s.Attempted, &s.Skipped, &s.PartialKeys, &s.PartialMistakes)
		if err != nil {
			return nil, err
		}
		s.Successful = s.Completed >= 3
		result = append(result, s)
	}
	return result, rows.Err()
}
func (db *Database) Stages(sid int64) ([]domain.StageResult, error) {
	rows, err := db.DB.Query(`SELECT c.id,COALESCE(c.file_path,''),COALESCE(c.start_line,0),COALESCE(c.end_line,0),COALESCE(c.language,''),c.code_content,COALESCE(c.comment_ranges,'[]'),COALESCE(c.difficulty_level,''),sr.keystrokes,sr.mistakes,sr.duration_ms,COALESCE(sr.wpm,0),COALESCE(sr.cpm,0),COALESCE(sr.accuracy,0),COALESCE(sr.score,0),COALESCE(sr.rank_name,''),COALESCE(sr.tier_name,''),COALESCE(sr.consistency_streaks,'[]'),sr.was_skipped,sr.was_failed FROM stages st JOIN challenges c ON c.id=st.challenge_id JOIN stage_results sr ON sr.stage_id=st.id WHERE st.session_id=? ORDER BY st.stage_number`, sid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.StageResult{}
	for rows.Next() {
		var s domain.StageResult
		var ranges, streaks string
		c := &s.Challenge
		err = rows.Scan(&c.ID, &c.Path, &c.StartLine, &c.EndLine, &c.Language, &c.Code, &ranges, &c.Difficulty, &s.Keystrokes, &s.Mistakes, &s.DurationMS, &s.WPM, &s.CPM, &s.Accuracy, &s.Score, &s.RankName, &s.TierName, &streaks, &s.Skipped, &s.Failed)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(ranges), &c.Comments); err != nil {
			return nil, fmt.Errorf("comment ranges: %w", err)
		}
		if err = json.Unmarshal([]byte(streaks), &s.Streaks); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func sqliteTimestamp(value string) string {
	if t, e := time.Parse(time.RFC3339Nano, value); e == nil {
		return t.UTC().Format("2006-01-02 15:04:05")
	}
	return value
}
func saveSessionMetadata(tx *sql.Tx, id int64, s domain.SessionResult) error {
	tp, tt, p, total := domain.RankPositions(s.Score)
	var best, worst domain.Metrics
	if len(s.Stages) > 0 {
		best = s.Stages[0].Metrics
		worst = best
	}
	for _, stage := range s.Stages {
		if stage.Score > best.Score {
			best = stage.Metrics
		}
		if stage.Score < worst.Score {
			worst = stage.Metrics
		}
	}
	_, err := tx.Exec("UPDATE session_results SET rank_position=?,rank_total=?,position=?,total=?,best_stage_wpm=?,worst_stage_wpm=?,best_stage_accuracy=?,worst_stage_accuracy=? WHERE session_id=?", tp, tt, p, total, best.WPM, worst.WPM, best.Accuracy, worst.Accuracy, id)
	return err
}
func (db *Database) Repositories() ([]domain.Repository, error) {
	rows, err := db.DB.Query("SELECT user_name,repository_name,remote_url FROM repositories ORDER BY user_name,repository_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Repository{}
	for rows.Next() {
		var r domain.Repository
		if err = rows.Scan(&r.Owner, &r.Name, &r.URL); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
