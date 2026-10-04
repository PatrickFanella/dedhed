package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/music"
	"github.com/jackc/pgx/v5"
)

var dateShape = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func nullDate(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// UpsertItems writes inventory rows, replacing earlier observations of the same item.
func (d *DB) UpsertItems(ctx context.Context, items []music.SourceItem) error {
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, it := range items {
		if it.ID == "" {
			return fmt.Errorf("item without identifier")
		}
		tracks, err := json.Marshal(it.Tracks)
		if err != nil {
			return err
		}
		warnings, _ := json.Marshal(it.Warnings)
		summary := it.Summary
		if len(summary) == 0 || !json.Valid(summary) {
			summary = []byte("{}")
		}
		var added any
		if !it.AddedAt.IsZero() {
			added = it.AddedAt
		}
		_, err = tx.Exec(ctx, `INSERT INTO source_items(id,show_date,date_text,artist,venue,location,title,kind,source,lineage,taper,transferer,downloads,avg_rating,num_reviews,item_size,added_at,tracks,songs,duration,warnings,summary,snapshot_sha256,observed_at,updated_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,now())
			ON CONFLICT(id) DO UPDATE SET show_date=excluded.show_date,date_text=excluded.date_text,artist=excluded.artist,venue=excluded.venue,location=excluded.location,title=excluded.title,kind=excluded.kind,source=excluded.source,lineage=excluded.lineage,taper=excluded.taper,transferer=excluded.transferer,downloads=excluded.downloads,avg_rating=excluded.avg_rating,num_reviews=excluded.num_reviews,item_size=excluded.item_size,added_at=excluded.added_at,tracks=excluded.tracks,songs=excluded.songs,duration=excluded.duration,warnings=excluded.warnings,summary=excluded.summary,snapshot_sha256=excluded.snapshot_sha256,observed_at=excluded.observed_at,updated_at=now()`,
			it.ID, nullDate(it.ShowDate), it.DateText, it.Artist, it.Venue, it.Location, it.Title, it.Kind, it.Source, it.Lineage, it.Taper, it.Transferer, it.Downloads, it.AvgRating, it.NumReviews, it.ItemSize, added, tracks, it.Songs, it.Duration, warnings, summary, it.Snapshot, it.ObservedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// YearCount is the number of items and distinct show dates in one year.
type YearCount struct {
	Year  int `json:"year"`
	Items int `json:"items"`
	Dates int `json:"dates"`
}

// DayCount is the number of items on one show date, with its venue for labels.
type DayCount struct {
	Date  string `json:"date"`
	Items int    `json:"items"`
	Venue string `json:"venue"`
}

func (d *DB) Years(ctx context.Context) ([]YearCount, error) {
	rows, err := d.Pool.Query(ctx, `SELECT EXTRACT(YEAR FROM show_date)::int, count(*)::int, count(DISTINCT show_date)::int FROM source_items WHERE `+d.publicPredicate()+` AND show_date IS NOT NULL GROUP BY 1 ORDER BY 1`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[YearCount])
}

// Days lists every show date in a year with its item count and most common venue.
func (d *DB) Days(ctx context.Context, year int) ([]DayCount, error) {
	if year < 1960 || year > 2100 {
		return nil, fmt.Errorf("year out of range")
	}
	rows, err := d.Pool.Query(ctx, `SELECT to_char(show_date,'YYYY-MM-DD'), count(*)::int, mode() WITHIN GROUP (ORDER BY venue) FROM source_items WHERE `+d.publicPredicate()+` AND show_date >= make_date($1,1,1) AND show_date < make_date($1+1,1,1) GROUP BY show_date ORDER BY show_date`, year)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[DayCount])
}

// OnThisDay lists every show date sharing a month and day, across years.
func (d *DB) OnThisDay(ctx context.Context, month, day int) ([]DayCount, error) {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return nil, fmt.Errorf("month or day out of range")
	}
	rows, err := d.Pool.Query(ctx, `SELECT to_char(show_date,'YYYY-MM-DD'), count(*)::int, mode() WITHIN GROUP (ORDER BY venue) FROM source_items WHERE `+d.publicPredicate()+` AND EXTRACT(MONTH FROM show_date)=$1 AND EXTRACT(DAY FROM show_date)=$2 GROUP BY show_date ORDER BY show_date`, month, day)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[DayCount])
}

const itemColumns = `id, coalesce(to_char(show_date,'YYYY-MM-DD'),''), date_text, artist, venue, location, title, kind, source, lineage, taper, transferer, downloads, avg_rating::float8, num_reviews, item_size, coalesce(added_at, 'epoch'::timestamptz), tracks, songs, duration, warnings, snapshot_sha256, observed_at`

func scanItem(rows pgx.CollectableRow) (music.SourceItem, error) {
	var it music.SourceItem
	var tracks, warnings []byte
	err := rows.Scan(&it.ID, &it.ShowDate, &it.DateText, &it.Artist, &it.Venue, &it.Location, &it.Title, &it.Kind, &it.Source, &it.Lineage, &it.Taper, &it.Transferer, &it.Downloads, &it.AvgRating, &it.NumReviews, &it.ItemSize, &it.AddedAt, &tracks, &it.Songs, &it.Duration, &warnings, &it.Snapshot, &it.ObservedAt)
	if err != nil {
		return it, err
	}
	if it.AddedAt.Unix() == 0 {
		it.AddedAt = time.Time{}
	}
	if err = json.Unmarshal(tracks, &it.Tracks); err != nil {
		return it, err
	}
	if it.Tracks == nil {
		it.Tracks = []music.Track{}
	}
	if err = json.Unmarshal(warnings, &it.Warnings); err != nil {
		return it, err
	}
	if it.Warnings == nil {
		it.Warnings = []string{}
	}
	return it, nil
}

// ItemsOn returns every item for one show date, best rated and most downloaded first.
func (d *DB) ItemsOn(ctx context.Context, date string) ([]music.SourceItem, error) {
	if !dateShape.MatchString(date) {
		return nil, fmt.Errorf("invalid date")
	}
	rows, err := d.Pool.Query(ctx, `SELECT `+itemColumns+` FROM source_items WHERE `+d.publicPredicate()+` AND show_date=$1::date ORDER BY avg_rating DESC, downloads DESC, id`, date)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanItem)
}

// SearchHit is one show date matched by a search, with its items.
type SearchHit struct {
	Date  string             `json:"date"`
	Venue string             `json:"venue"`
	Items []music.SourceItem `json:"items"`
}

// Search runs a full-text query over titles, venues, places, credits and song
// titles, grouping hits by show date. Dates without a month or day are excluded.
func (d *DB) Search(ctx context.Context, q, year string, page, size int) ([]SearchHit, int, error) {
	if page < 1 || size < 1 || size > 100 {
		return nil, 0, fmt.Errorf("invalid page")
	}
	words := []string{}
	for _, w := range strings.Fields(strings.ToLower(q)) {
		w = regexp.MustCompile(`[^\p{L}\p{N}]+`).ReplaceAllString(w, "")
		if w != "" {
			words = append(words, w+":*")
		}
	}
	if len(words) > 12 {
		return nil, 0, fmt.Errorf("query too long")
	}
	query := strings.Join(words, " & ")
	where := d.publicPredicate() + ` AND show_date IS NOT NULL`
	args := []any{}
	if query != "" {
		args = append(args, query)
		where += fmt.Sprintf(` AND search @@ to_tsquery('simple', $%d)`, len(args))
	}
	if year != "" {
		if !regexp.MustCompile(`^\d{4}$`).MatchString(year) {
			return nil, 0, fmt.Errorf("invalid year")
		}
		args = append(args, year)
		where += fmt.Sprintf(` AND EXTRACT(YEAR FROM show_date)=$%d::int`, len(args))
	}
	var total int
	if err := d.Pool.QueryRow(ctx, `SELECT count(DISTINCT show_date) FROM source_items WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, size, (page-1)*size)
	rows, err := d.Pool.Query(ctx, `SELECT `+itemColumns+` FROM source_items WHERE show_date IN (SELECT show_date FROM source_items WHERE `+where+fmt.Sprintf(` GROUP BY show_date ORDER BY show_date LIMIT $%d OFFSET $%d) AND `, len(args)-1, len(args))+where+` ORDER BY show_date, avg_rating DESC, downloads DESC, id`, args...)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, scanItem)
	if err != nil {
		return nil, 0, err
	}
	hits := []SearchHit{}
	for _, it := range items {
		if len(hits) == 0 || hits[len(hits)-1].Date != it.ShowDate {
			hits = append(hits, SearchHit{Date: it.ShowDate, Venue: it.Venue})
		}
		hits[len(hits)-1].Items = append(hits[len(hits)-1].Items, it)
	}
	return hits, total, nil
}

// ItemCount reports how many items are indexed and how many have fetched metadata.
func (d *DB) ItemCount(ctx context.Context) (int, int, error) {
	var items, fetched int
	err := d.Pool.QueryRow(ctx, `SELECT count(*)::int, count(*) FILTER (WHERE snapshot_sha256<>'')::int FROM source_items`).Scan(&items, &fetched)
	return items, fetched, err
}
