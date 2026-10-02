package catalog

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/music"
)

func testDB(t *testing.T) (*DB, context.Context) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires an explicitly disposable TEST_DATABASE_URL")
	}
	ctx := context.Background()
	admin, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Pool.Close)
	schema := "test_" + music.Identity(t.Name(), time.Now().String())
	if _, err = admin.Pool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = admin.Pool.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE") })
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := Open(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Pool.Close)
	if err = db.Migrate(ctx, "../../../migrations"); err != nil {
		t.Fatal(err)
	}
	return db, ctx
}

func TestInventoryQueries(t *testing.T) {
	db, ctx := testDB(t)
	now := time.Now().UTC()
	items := []music.SourceItem{
		{ID: "gd77-05-08.aud", ShowDate: "1977-05-08", DateText: "1977-05-08", Venue: "Barton Hall", Location: "Ithaca, NY", Title: "Grateful Dead Live at Barton Hall on 1977-05-08", Kind: "audience", Taper: "Jerry Moore", AvgRating: 4.8, Downloads: 900, Tracks: []music.Track{{Title: "Scarlet Begonias", Duration: 10}, {Title: "Fire On The Mountain", Start: 10, Duration: 5}}, Songs: "Scarlet Begonias / Fire On The Mountain", Duration: 15, Warnings: []string{}, Snapshot: "abc", ObservedAt: now},
		{ID: "gd77-05-08.sbd", ShowDate: "1977-05-08", DateText: "1977-05-08", Venue: "Barton Hall", Location: "Ithaca, NY", Title: "Barton Hall soundboard", Kind: "soundboard", AvgRating: 5, Downloads: 5000, Tracks: []music.Track{}, Warnings: []string{"Full metadata not fetched"}, ObservedAt: now},
		{ID: "gd89-07-07", ShowDate: "1989-07-07", DateText: "1989-07-07", Venue: "JFK Stadium", Location: "Philadelphia, PA", Title: "Grateful Dead Live at JFK Stadium on 1989-07-07", Kind: "audience", Tracks: []music.Track{}, Songs: "Touch Of Grey", Warnings: []string{}, ObservedAt: now},
		{ID: "gd68-00-00", ShowDate: "", DateText: "1968-00-00", Venue: "Unknown", Title: "Undated 1968 tape", Kind: "unknown", Tracks: []music.Track{}, Warnings: []string{}, ObservedAt: now},
		{ID: "gd91-05-08", ShowDate: "1991-05-08", DateText: "1991-05-08", Venue: "Shoreline", Location: "Mountain View, CA", Title: "Grateful Dead Live at Shoreline on 1991-05-08", Kind: "audience", Tracks: []music.Track{}, Warnings: []string{}, ObservedAt: now},
	}
	if err := db.UpsertItems(ctx, items); err != nil {
		t.Fatal(err)
	}
	// Re-observing an item replaces it rather than duplicating it.
	items[0].Downloads = 901
	if err := db.UpsertItems(ctx, items[:1]); err != nil {
		t.Fatal(err)
	}
	total, fetched, err := db.ItemCount(ctx)
	if err != nil || total != 5 || fetched != 1 {
		t.Fatalf("%d %d %v", total, fetched, err)
	}
	years, err := db.Years(ctx)
	if err != nil || len(years) != 3 || years[0].Year != 1977 || years[0].Items != 2 || years[0].Dates != 1 {
		t.Fatalf("%+v %v", years, err)
	}
	days, err := db.Days(ctx, 1977)
	if err != nil || len(days) != 1 || days[0].Date != "1977-05-08" || days[0].Items != 2 || days[0].Venue != "Barton Hall" {
		t.Fatalf("%+v %v", days, err)
	}
	if _, err = db.Days(ctx, 1900); err == nil {
		t.Fatal("year bound")
	}
	today, err := db.OnThisDay(ctx, 5, 8)
	if err != nil || len(today) != 2 || today[0].Date != "1977-05-08" || today[1].Date != "1991-05-08" {
		t.Fatalf("%+v %v", today, err)
	}
	on, err := db.ItemsOn(ctx, "1977-05-08")
	if err != nil || len(on) != 2 || on[0].ID != "gd77-05-08.sbd" || on[1].Downloads != 901 || len(on[1].Tracks) != 2 || on[1].Tracks[1].Title != "Fire On The Mountain" {
		t.Fatalf("%+v %v", on, err)
	}
	if _, err = db.ItemsOn(ctx, "1977-5-8"); err == nil {
		t.Fatal("date shape")
	}
	// Song titles, venues and credits are searchable; prefixes match; undated items are excluded.
	for q, want := range map[string]string{"fire mountain": "1977-05-08", "touch": "1989-07-07", "jerry moore": "1977-05-08", "shore": "1991-05-08", "undated": ""} {
		hits, n, err := db.Search(ctx, q, "", 1, 20)
		if err != nil {
			t.Fatal(q, err)
		}
		if want == "" {
			if n != 0 {
				t.Fatalf("%q matched %+v", q, hits)
			}
			continue
		}
		if n != 1 || len(hits) != 1 || hits[0].Date != want {
			t.Fatalf("%q: %d %+v", q, n, hits)
		}
	}
	hits, n, err := db.Search(ctx, "", "1977", 1, 20)
	if err != nil || n != 1 || len(hits[0].Items) != 2 {
		t.Fatalf("%d %+v %v", n, hits, err)
	}
	all, n, err := db.Search(ctx, "", "", 2, 2)
	if err != nil || n != 3 || len(all) != 1 || all[0].Date != "1991-05-08" {
		t.Fatalf("page 2: %d %+v %v", n, all, err)
	}
	if _, _, err = db.Search(ctx, "", "77", 1, 20); err == nil {
		t.Fatal("year shape")
	}
}
