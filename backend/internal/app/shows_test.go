package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/catalog"
	"git.subcult.tv/subculture-collective/dedhed/internal/fixture"
	"git.subcult.tv/subculture-collective/dedhed/internal/music"
)

func TestShowIndexEndpoints(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable TEST_DATABASE_URL")
	}
	ctx := context.Background()
	admin, err := catalog.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Pool.Close()
	schema := "test_" + music.Identity("shows", time.Now().String())
	if _, err = admin.Pool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Pool.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
	u, _ := url.Parse(dsn)
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := catalog.Open(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Pool.Close()
	if err = db.Migrate(ctx, "../../../migrations"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = fixture.Seed(ctx, db, dir); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	// The seeded synthetic-1977 recording is playable; the same item id in the inventory must report that.
	if err = db.UpsertItems(ctx, []music.SourceItem{
		{ID: "synthetic-1977", ShowDate: "1977-05-08", DateText: "1977-05-08", Venue: "Barton Hall (synthetic)", Kind: "synthetic", Tracks: []music.Track{}, Warnings: []string{}, ObservedAt: now},
		{ID: "gd77-05-08.other", ShowDate: "1977-05-08", DateText: "1977-05-08", Venue: "Barton Hall", Kind: "soundboard", AvgRating: 5, Tracks: []music.Track{}, Warnings: []string{}, ObservedAt: now},
		{ID: "gd80-05-08", ShowDate: "1980-05-08", DateText: "1980-05-08", Venue: "Folsom Field", Kind: "audience", Tracks: []music.Track{}, Warnings: []string{}, ObservedAt: now},
	}); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	fixed := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer((&Server{DB: db, Root: root, Now: func() time.Time { return fixed }}).Handler())
	defer srv.Close()
	get := func(path string, want int) map[string]any {
		t.Helper()
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != want {
			t.Fatalf("%s: %d", path, resp.StatusCode)
		}
		var out map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&out)
		return out
	}
	cal := get("/api/calendar?year=1977", 200)
	if len(cal["years"].([]any)) != 2 || len(cal["days"].([]any)) != 1 {
		t.Fatalf("%+v", cal)
	}
	get("/api/calendar?year=abc", 400)
	day := get("/api/dates/1977-05-08", 200)
	items := day["items"].([]any)
	if day["playable"].(float64) != 1 || len(items) != 2 || items[0].(map[string]any)["library"] != "playable" || items[1].(map[string]any)["library"] != "" {
		t.Fatalf("%+v", day)
	}
	get("/api/dates/1977-5-8", 400)
	today := get("/api/on-this-day", 200)
	if today["month"].(float64) != 5 || len(today["dates"].([]any)) != 2 {
		t.Fatalf("%+v", today)
	}
	if len(get("/api/on-this-day?month=2&day=30", 200)["dates"].([]any)) != 0 {
		t.Fatal("february 30")
	}
	get("/api/on-this-day?month=13", 400)
	found := get("/api/shows?q=folsom", 200)
	if found["total"].(float64) != 1 || len(found["dates"].([]any)) != 1 {
		t.Fatalf("%+v", found)
	}
	get("/api/shows?year=19", 400)
}
