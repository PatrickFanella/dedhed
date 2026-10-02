// Package inventory builds the collection-wide source index: a scrape of the
// Internet Archive collection listing, a resumable per-item metadata fetch, and
// normalization into music.SourceItem rows. It runs as a bounded operator
// workflow outside isolated development networks and never acquires audio.
package inventory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/music"
	"git.subcult.tv/subculture-collective/dedhed/internal/source"
	ia "git.subcult.tv/subculture-collective/dedhed/third_party/arkhiv/internetarchive"
)

// Fields requested from the scrape endpoint. Descriptions are omitted: they are
// large, free-form, and the full metadata fetch preserves them anyway.
const Fields = "identifier,mediatype,creator,date,title,addeddate,venue,coverage,source,lineage,taper,transferer,item_size,downloads,avg_rating,num_reviews"

var validID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,200}$`)

// Summary is one record from the collection scrape.
type Summary struct {
	Identifier string  `json:"identifier"`
	MediaType  string  `json:"mediatype"`
	Creator    any     `json:"creator"`
	Date       string  `json:"date"`
	Title      string  `json:"title"`
	AddedDate  string  `json:"addeddate"`
	Venue      any     `json:"venue"`
	Coverage   any     `json:"coverage"`
	Source     any     `json:"source"`
	Lineage    any     `json:"lineage"`
	Taper      any     `json:"taper"`
	Transferer any     `json:"transferer"`
	ItemSize   int64   `json:"item_size"`
	Downloads  int     `json:"downloads"`
	AvgRating  float64 `json:"avg_rating"`
	NumReviews int     `json:"num_reviews"`
}

func text(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case []any:
		parts := []string{}
		for _, v := range x {
			if s := text(v); s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, "; ")
	default:
		return ""
	}
}

// Scrape pages through the collection listing and writes one JSON line per item
// to w. It returns the number of items written and the total the source reported.
func Scrape(ctx context.Context, c *ia.Client, query string, w io.Writer) (int, int, error) {
	cursor, written, total := "", 0, 0
	for page := 0; page < 100; page++ {
		q := url.Values{"q": {query}, "fields": {Fields}, "count": {"10000"}}
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		resp, err := c.Request(ctx, "/services/search/v1/scrape?"+q.Encode(), nil)
		if err != nil {
			return written, total, err
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
		resp.Body.Close()
		if err != nil {
			return written, total, err
		}
		if resp.StatusCode != 200 {
			return written, total, fmt.Errorf("scrape status %d", resp.StatusCode)
		}
		var data struct {
			Items  []json.RawMessage `json:"items"`
			Cursor string            `json:"cursor"`
			Total  int               `json:"total"`
			Error  string            `json:"error"`
		}
		if err = json.Unmarshal(body, &data); err != nil {
			return written, total, err
		}
		if data.Error != "" {
			return written, total, fmt.Errorf("scrape: %s", data.Error)
		}
		if data.Total > total { // the source reports the remaining count per page; keep the first, largest figure
			total = data.Total
		}
		for _, item := range data.Items {
			var s Summary
			if err = json.Unmarshal(item, &s); err != nil || !validID.MatchString(s.Identifier) {
				continue
			}
			if _, err = w.Write(append(item, '\n')); err != nil {
				return written, total, err
			}
			written++
		}
		if data.Cursor == "" || len(data.Items) == 0 {
			return written, total, nil
		}
		cursor = data.Cursor
	}
	return written, total, fmt.Errorf("scrape paging exceeded bounds")
}

// ReadSummaries reads the JSON-lines file written by Scrape.
func ReadSummaries(path string) ([]Summary, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := []Summary{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var s Summary
		if err := json.Unmarshal(sc.Bytes(), &s); err != nil {
			return nil, err
		}
		if validID.MatchString(s.Identifier) {
			out = append(out, s)
		}
	}
	return out, sc.Err()
}

// Progress is reported after each item during FetchAll.
type Progress struct {
	Done, Skipped, Failed, Total int
	Last                         string
}

// FetchAll stores each item's full metadata as items/<id>.json under dir. Existing
// files are kept, so an interrupted run resumes. Failures are appended to
// errors.jsonl and do not stop the run. Concurrency and delay bound the load.
func FetchAll(ctx context.Context, a *source.Adapter, ids []string, dir string, concurrency int, delay time.Duration, report func(Progress)) (Progress, error) {
	if concurrency < 1 || concurrency > 4 {
		return Progress{}, fmt.Errorf("concurrency must be 1–4")
	}
	items := filepath.Join(dir, "items")
	if err := os.MkdirAll(items, 0o700); err != nil {
		return Progress{}, err
	}
	errLog, err := os.OpenFile(filepath.Join(dir, "errors.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return Progress{}, err
	}
	defer errLog.Close()
	var mu sync.Mutex
	p := Progress{Total: len(ids)}
	work := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range work {
				target := filepath.Join(items, id+".json")
				if _, statErr := os.Stat(target); statErr == nil {
					mu.Lock()
					p.Skipped++
					p.Last = id
					report(p)
					mu.Unlock()
					continue
				}
				_, raw, fetchErr := a.Fetch(ctx, id)
				mu.Lock()
				if fetchErr != nil && len(raw) == 0 {
					p.Failed++
					_, _ = fmt.Fprintf(errLog, "%s\n", mustJSON(map[string]any{"item": id, "error": fetchErr.Error(), "at": time.Now().UTC()}))
				} else {
					// Normalization problems are recorded at import time; the raw snapshot is still worth keeping.
					tmp := target + ".part"
					if writeErr := os.WriteFile(tmp, raw, 0o600); writeErr != nil || os.Rename(tmp, target) != nil {
						p.Failed++
					} else {
						p.Done++
					}
				}
				p.Last = id
				report(p)
				mu.Unlock()
				if delay > 0 {
					select {
					case <-ctx.Done():
					case <-time.After(delay):
					}
				}
			}
		}()
	}
	for _, id := range ids {
		if ctx.Err() != nil {
			break
		}
		if validID.MatchString(id) {
			work <- id
		}
	}
	close(work)
	wg.Wait()
	return p, ctx.Err()
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

// ShowDate reduces a source date such as "1977-05-08T00:00:00Z" or "1968-00-00"
// to YYYY-MM-DD. An unknown month or day yields "".
func ShowDate(v string) string {
	v = strings.TrimSpace(v)
	if len(v) < 10 {
		return ""
	}
	v = v[:10]
	t, err := time.Parse("2006-01-02", v)
	if err != nil || t.Year() < 1960 || t.Year() > 2100 {
		return ""
	}
	return v
}

// Normalize combines a scrape summary with the item's full metadata (which may be
// nil when the fetch has not happened yet) into a SourceItem.
func Normalize(s Summary, raw []byte, observed time.Time) music.SourceItem {
	item := music.SourceItem{ID: s.Identifier, ShowDate: ShowDate(s.Date), DateText: strings.TrimSpace(s.Date), Artist: text(s.Creator), Venue: text(s.Venue), Location: text(s.Coverage), Title: strings.TrimSpace(s.Title), Kind: "unknown", Source: text(s.Source), Lineage: text(s.Lineage), Taper: text(s.Taper), Transferer: text(s.Transferer), Downloads: s.Downloads, AvgRating: s.AvgRating, NumReviews: s.NumReviews, ItemSize: s.ItemSize, Tracks: []music.Track{}, Warnings: []string{}, ObservedAt: observed}
	if len(item.DateText) >= 10 {
		item.DateText = item.DateText[:10]
	}
	if t, err := time.Parse(time.RFC3339, s.AddedDate); err == nil {
		item.AddedAt = t
	}
	item.Summary, _ = json.Marshal(s)
	if len(raw) == 0 {
		item.Warnings = append(item.Warnings, "Full metadata not fetched")
		item.Kind = classify(item.Source)
		return item
	}
	item.Snapshot = music.Digest(raw)
	r, err := source.Normalize(s.Identifier, raw)
	if r.ID == "" {
		item.Warnings = append(item.Warnings, "Metadata unreadable: "+err.Error())
		item.Kind = classify(item.Source)
		return item
	}
	if err != nil {
		item.Warnings = append(item.Warnings, "Normalization: "+err.Error())
	}
	item.Kind = r.Kind
	if r.Date != "" {
		item.DateText = r.Date
		if len(item.DateText) >= 10 {
			item.DateText = item.DateText[:10]
		}
		if d := ShowDate(r.Date); d != "" {
			item.ShowDate = d
		}
	}
	for field, value := range map[*string]string{&item.Artist: r.Artist, &item.Venue: r.Venue, &item.Location: r.Location, &item.Title: r.Title, &item.Source: r.Source, &item.Taper: r.Taper, &item.Transferer: r.Transferer} {
		if value != "" {
			*field = value
		}
	}
	item.Tracks = r.Tracks
	item.Duration = r.Duration
	titles := []string{}
	for _, t := range r.Tracks {
		titles = append(titles, t.Title)
	}
	item.Songs = strings.Join(titles, " / ")
	item.Warnings = append(item.Warnings, r.Warnings...)
	return item
}

func classify(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "matrix") || strings.Contains(s, "mtx") || (strings.Contains(s, "sbd") && strings.Contains(s, "aud")):
		return "matrix"
	case strings.Contains(s, "soundboard") || strings.Contains(s, "sbd"):
		return "soundboard"
	case strings.Contains(s, "aud"):
		return "audience"
	}
	return "unknown"
}

// Load reads the scrape file and any fetched metadata under dir, calling emit for
// each normalized item in batches of at most size.
func Load(dir string, size int, emit func([]music.SourceItem) error) (int, int, error) {
	summaries, err := ReadSummaries(filepath.Join(dir, "collection.jsonl"))
	if err != nil {
		return 0, 0, err
	}
	observed := time.Now().UTC()
	if st, err := os.Stat(filepath.Join(dir, "collection.jsonl")); err == nil {
		observed = st.ModTime().UTC()
	}
	batch := []music.SourceItem{}
	fetched, total := 0, 0
	for _, s := range summaries {
		raw, _ := os.ReadFile(filepath.Join(dir, "items", s.Identifier+".json"))
		if len(raw) > 0 {
			fetched++
		}
		batch = append(batch, Normalize(s, raw, observed))
		total++
		if len(batch) >= size {
			if err := emit(batch); err != nil {
				return total, fetched, err
			}
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		if err := emit(batch); err != nil {
			return total, fetched, err
		}
	}
	return total, fetched, nil
}
