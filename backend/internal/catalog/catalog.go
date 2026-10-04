package catalog

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"git.subcult.tv/subculture-collective/dedhed/internal/music"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool   *pgxpool.Pool
	Public bool
}

func Open(ctx context.Context, url string) (*DB, error) {
	p, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err = p.Ping(ctx); err != nil {
		p.Close()
		return nil, err
	}
	return &DB{Pool: p}, nil
}
func (d *DB) Migrate(ctx context.Context, dir string) error {
	c, err := d.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer c.Release()
	if _, err = c.Exec(ctx, `SELECT pg_advisory_lock(78439211)`); err != nil {
		return err
	}
	defer c.Exec(context.Background(), `SELECT pg_advisory_unlock(78439211)`)
	if _, err = c.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_versions (name text PRIMARY KEY,sha256 text NOT NULL)`); err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no migration files")
	}
	sort.Strings(files)
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		name := filepath.Base(f)
		hash := music.Digest(b)
		var old string
		err = c.QueryRow(ctx, `SELECT sha256 FROM schema_versions WHERE name=$1`, name).Scan(&old)
		if err == nil {
			if old != hash {
				return fmt.Errorf("migration checksum changed: %s", name)
			}
			continue
		}
		if err != pgx.ErrNoRows {
			return err
		}
		tx, err := c.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, string(b)); err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO schema_versions VALUES($1,$2)`, name, hash)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}
func (d *DB) Import(ctx context.Context, r music.Recording, raw []byte) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if !json.Valid(raw) {
		return fmt.Errorf("invalid raw snapshot")
	}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	hash := music.Digest(raw)
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `INSERT INTO source_snapshots(sha256,item_id,raw) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, hash, r.ID, raw); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO recordings(id,source_hash,source) VALUES($1,$2,$3) ON CONFLICT(id) DO UPDATE SET source_hash=excluded.source_hash,source=excluded.source,updated_at=now()`, r.ID, hash, b); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (d *DB) Entries(ctx context.Context) (map[string]music.Entry, error) {
	rows, err := d.Pool.Query(ctx, `SELECT source,review,media,media_sha256,source_hash,playback FROM recordings WHERE `+d.publicPredicate()+` ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]music.Entry{}
	for rows.Next() {
		var e music.Entry
		var source, review, playback []byte
		if err = rows.Scan(&source, &review, &e.Media, &e.SHA256, &e.SourceHash, &playback); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(source, &e.Recording); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(review, &e.Review); err != nil {
			return nil, err
		}
		if len(playback) > 0 {
			if err = json.Unmarshal(playback, &e.Recording); err != nil {
				return nil, err
			}
		}
		if e.Review.Kind != "" {
			e.Recording.Kind = e.Review.Kind
		}
		if len(e.Review.Sets) > 0 {
			e.Recording.Sets = e.Review.Sets
		}
		if e.Review.Title != "" {
			e.Recording.Title = e.Review.Title
		}
		if e.Review.ConcertID != "" {
			e.Recording.ConcertID = e.Review.ConcertID
		}
		if d.Public {
			e.Review.PrivateOnly = false
			e.Review.Relay = true
			e.Review.Acquisition = false
			e.Review.Notes = "Approved noncommercial audience-recording playback and scheduled relay. Source warnings and credits remain with the recording."
		}
		out[e.Recording.ID] = e
	}
	return out, rows.Err()
}
func (d *DB) Review(ctx context.Context, id string, review music.Review) error {
	if review.ConcertID != "" && !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`).MatchString(review.ConcertID) {
		return fmt.Errorf("invalid reviewed concert identity")
	}
	entries, err := d.Entries(ctx)
	if err != nil {
		return err
	}
	e, ok := entries[id]
	if !ok {
		return fmt.Errorf("recording missing")
	}
	if len(review.Sets) > 0 {
		if err = music.ValidateSets(review.Sets, len(e.Recording.Tracks)); err != nil {
			return err
		}
	}
	if (review.Acquisition || review.Playback || review.Relay) && !e.Recording.Synthetic && review.Kind != "audience" {
		return fmt.Errorf("pilot requires reviewed audience source")
	}
	if review.Relay {
		return fmt.Errorf("relay not enabled in private pilot")
	}
	if review.Playback && !review.PrivateOnly {
		return fmt.Errorf("only private playback supported")
	}
	if review.Published && (!review.Playback || e.Media == "") {
		return fmt.Errorf("publication requires verified media and playback decision")
	}
	b, err := json.Marshal(review)
	if err != nil {
		return err
	}
	_, err = d.Pool.Exec(ctx, `UPDATE recordings SET review=$2,updated_at=now() WHERE id=$1`, id, b)
	return err
}
func (d *DB) Media(ctx context.Context, id, name, sha string, r *music.Recording) error {
	_, hashErr := hex.DecodeString(sha)
	if len(sha) != 64 || hashErr != nil || name == "." || name == "" || filepath.Base(name) != name {
		return fmt.Errorf("invalid media identity")
	}
	if r != nil {
		if r.ID != id {
			return fmt.Errorf("playback identity differs from recording")
		}
		if err := r.Validate(); err != nil {
			return err
		}
		b, err := json.Marshal(r)
		if err != nil {
			return err
		}
		result, err := d.Pool.Exec(ctx, `UPDATE recordings SET media=$2,media_sha256=$3,playback=$4,updated_at=now() WHERE id=$1`, id, name, sha, b)
		if err == nil && result.RowsAffected() != 1 {
			return fmt.Errorf("recording missing")
		}
		return err
	}
	result, err := d.Pool.Exec(ctx, `UPDATE recordings SET media=$2,media_sha256=$3,playback=NULL,updated_at=now() WHERE id=$1`, id, name, sha)
	if err == nil && result.RowsAffected() != 1 {
		return fmt.Errorf("recording missing")
	}
	return err
}
func (d *DB) SaveSchedule(ctx context.Context, s music.Schedule) error {
	if err := s.Validate(); err != nil {
		return err
	}
	entries, err := d.Entries(ctx)
	if err != nil {
		return err
	}
	for _, p := range s.Programs {
		e, ok := entries[p.RecordingID]
		if !ok || !e.Review.Published || !e.Review.Playback || e.Media == "" || p.Duration > e.Recording.Duration {
			return fmt.Errorf("schedule references ineligible recording")
		}
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = d.Pool.Exec(ctx, `INSERT INTO station_schedule(id,schedule) VALUES(1,$1) ON CONFLICT(id) DO UPDATE SET schedule=excluded.schedule,updated_at=now()`, b)
	return err
}
func (d *DB) Schedule(ctx context.Context) (music.Schedule, error) {
	var b []byte
	err := d.Pool.QueryRow(ctx, `SELECT schedule FROM station_schedule WHERE id=1`).Scan(&b)
	if err != nil {
		return music.Schedule{}, err
	}
	var s music.Schedule
	err = json.Unmarshal(b, &s)
	return s, err
}
