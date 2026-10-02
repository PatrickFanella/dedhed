package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/music"
	"git.subcult.tv/subculture-collective/dedhed/internal/source"
	ia "git.subcult.tv/subculture-collective/dedhed/third_party/arkhiv/internetarchive"
)

func TestShowDate(t *testing.T) {
	for in, want := range map[string]string{"1977-05-08T00:00:00Z": "1977-05-08", "1977-05-08": "1977-05-08", "1968-00-00": "", "1968-07-00T00:00:00Z": "", "": "", "1900-01-01": "", "garbage": ""} {
		if got := ShowDate(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestScrapePagesUntilCursorEnds(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/services/search/v1/scrape" || r.URL.Query().Get("count") != "10000" {
			t.Errorf("unexpected request %s", r.URL)
		}
		if r.URL.Query().Get("cursor") == "" {
			_, _ = w.Write([]byte(`{"items":[{"identifier":"gd77-05-08","date":"1977-05-08T00:00:00Z","venue":"Barton Hall"},{"identifier":"../bad"}],"cursor":"next","total":3}`))
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"identifier":"gd77-05-09","date":"1977-05-09T00:00:00Z"}],"total":3}`))
	}))
	defer srv.Close()
	c := ia.New()
	c.Base = srv.URL
	var out bytes.Buffer
	written, total, err := Scrape(context.Background(), c, "collection:GratefulDead", &out)
	if err != nil || written != 2 || total != 3 || calls != 2 {
		t.Fatalf("written=%d total=%d calls=%d err=%v", written, total, calls, err)
	}
	if strings.Count(out.String(), "\n") != 2 || strings.Contains(out.String(), "bad") {
		t.Fatal(out.String())
	}
}

func TestNormalizeWithAndWithoutMetadata(t *testing.T) {
	s := Summary{Identifier: "gd1977-05-08.test", Date: "1977-05-08T00:00:00Z", Title: "Grateful Dead Live at Barton Hall on 1977-05-08", Venue: "Barton Hall", Coverage: "Ithaca, NY", Source: "AUD > cassette", Taper: []any{"Jerry Moore"}, AvgRating: 4.5, Downloads: 12}
	item := Normalize(s, nil, time.Unix(0, 0).UTC())
	if item.ShowDate != "1977-05-08" || item.Kind != "audience" || item.Taper != "Jerry Moore" || len(item.Tracks) != 0 || item.Snapshot != "" || item.Warnings[0] != "Full metadata not fetched" {
		t.Fatalf("%+v", item)
	}
	raw := []byte(`{"metadata":{"creator":"Grateful Dead","date":"1977-05-08","venue":"Barton Hall","coverage":"Ithaca, NY","source":"SBD","title":"GD 1977-05-08"},"files":[{"name":"t1.mp3","format":"VBR MP3","title":"Scarlet Begonias","length":"10","size":"100"},{"name":"t2.mp3","format":"VBR MP3","title":"Fire On The Mountain","length":"5","size":"100"}]}`)
	item = Normalize(s, raw, time.Unix(0, 0).UTC())
	if item.Kind != "soundboard" || len(item.Tracks) != 2 || item.Duration != 15 || item.Songs != "Scarlet Begonias / Fire On The Mountain" || item.Snapshot == "" || item.Title != "GD 1977-05-08" {
		t.Fatalf("%+v", item)
	}
	// Unreadable metadata keeps the summary and records the problem instead of dropping the item.
	item = Normalize(s, []byte(`{"files":[]}`), time.Unix(0, 0).UTC())
	if item.ShowDate != "1977-05-08" || len(item.Warnings) == 0 || !strings.HasPrefix(item.Warnings[0], "Metadata unreadable") {
		t.Fatalf("%+v", item)
	}
	// A bad track duration keeps the readable tracks and records the normalization error.
	item = Normalize(s, []byte(`{"metadata":{"date":"1977-05-08"},"files":[{"name":"t1.mp3","format":"VBR MP3","length":"NaN"}]}`), time.Unix(0, 0).UTC())
	found := false
	for _, w := range item.Warnings {
		found = found || strings.HasPrefix(w, "Normalization:")
	}
	if !found {
		t.Fatalf("%+v", item.Warnings)
	}
}

func TestFetchAllResumesAndRecordsFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/broken") {
			http.Error(w, "nope", 404)
			return
		}
		_, _ = w.Write([]byte(`{"metadata":{"date":"1977-05-08"},"files":[{"name":"t1.mp3","format":"VBR MP3","length":"10"}]}`))
	}))
	defer srv.Close()
	a := source.New()
	a.Client.Base = srv.URL
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "items"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "items", "already.json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	p, err := FetchAll(context.Background(), a, []string{"already", "fresh", "broken"}, dir, 2, 0, func(Progress) {})
	if err != nil || p.Skipped != 1 || p.Done != 1 || p.Failed != 1 {
		t.Fatalf("%+v %v", p, err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "items", "already.json")); string(b) != "{}" {
		t.Fatal("existing snapshot was overwritten")
	}
	if _, err := os.Stat(filepath.Join(dir, "items", "fresh.json")); err != nil {
		t.Fatal(err)
	}
	errs, _ := os.ReadFile(filepath.Join(dir, "errors.jsonl"))
	var rec map[string]any
	if json.Unmarshal(bytes.TrimSpace(errs), &rec) != nil || rec["item"] != "broken" {
		t.Fatalf("%s", errs)
	}
	// Load pairs summaries with fetched snapshots and reports both counts.
	if err := os.WriteFile(filepath.Join(dir, "collection.jsonl"), []byte("{\"identifier\":\"fresh\",\"date\":\"1977-05-08\"}\n{\"identifier\":\"missing\",\"date\":\"1977-05-09\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	batches, items := 0, 0
	total, fetched, err := Load(dir, 1, func(b []music.SourceItem) error { batches++; items += len(b); return nil })
	if err != nil || total != 2 || fetched != 1 || batches != 2 || items != 2 {
		t.Fatalf("total=%d fetched=%d batches=%d items=%d err=%v", total, fetched, batches, items, err)
	}
}
