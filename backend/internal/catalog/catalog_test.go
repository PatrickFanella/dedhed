package catalog

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/music"
)

func TestImportReviewAndRecoveryState(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("requires an explicitly disposable TEST_DATABASE_URL")
	}
	ctx := context.Background()
	admin, err := Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Pool.Close()
	schema := "test_" + music.Identity(time.Now().String())
	if _, err = admin.Pool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Pool.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	d := &DB{Pool: pool}

	if err = d.Migrate(ctx, "../../../migrations"); err != nil {
		t.Fatal(err)
	}
	if err = d.Migrate(ctx, "../../../migrations"); err != nil {
		t.Fatal(err)
	}
	r := music.Recording{ID: "aud", ConcertID: "show", Kind: "audience", Duration: 10, Tracks: []music.Track{{Position: 0, Duration: 10}}, Sets: []music.Set{}}
	for range 2 {
		if err = d.Import(ctx, r, []byte(`{"raw":1}`)); err != nil {
			t.Fatal(err)
		}
	}
	if err = d.Review(ctx, r.ID, music.Review{Kind: "audience", Playback: true, PrivateOnly: true, Published: true}); err == nil {
		t.Fatal("published before media")
	}
	if err = d.Media(ctx, r.ID, "media.wav", music.Digest([]byte("audio")), nil); err != nil {
		t.Fatal(err)
	}

	wrong := r
	wrong.ID = "other"
	if err = d.Media(ctx, r.ID, "wrong.wav", music.Digest([]byte("wrong")), &wrong); err == nil {
		t.Fatal("attached playback metadata to another recording")
	}
	if err = d.Media(ctx, "missing", "media.wav", music.Digest([]byte("audio")), nil); err == nil {
		t.Fatal("attached media to nonexistent recording")
	}
	measured := r
	measured.Tracks = append([]music.Track(nil), r.Tracks...)
	measured.Duration = 12
	measured.Tracks[0].Duration = 12
	if err = d.Media(ctx, r.ID, "measured.wav", music.Digest([]byte("measured")), &measured); err != nil {
		t.Fatal(err)
	}
	if err = d.Media(ctx, r.ID, "media.wav", music.Digest([]byte("audio")), nil); err != nil {
		t.Fatal(err)
	}
	reset, err := d.Entries(ctx)
	if err != nil || reset[r.ID].Recording.Duration != r.Duration {
		t.Fatal("replacing media retained unrelated measured timing", reset, err)
	}
	if err = d.Review(ctx, r.ID, music.Review{ConcertID: "invalid group"}); err == nil {
		t.Fatal("accepted invalid reviewed concert identity")
	}
	review := music.Review{Kind: "audience", Published: true, Playback: true, PrivateOnly: true, Title: "Editorial title", ConcertID: "reviewed-show", Sets: []music.Set{{Title: "Set I", First: 0, Last: 0, Reviewed: true}}}
	if err = d.Review(ctx, r.ID, review); err != nil {
		t.Fatal(err)
	}
	r.Title = "New source title"
	r.ConcertID = "changed-source-group"
	if err = d.Import(ctx, r, []byte(`{"raw":2}`)); err != nil {
		t.Fatal(err)
	}
	entries, err := d.Entries(ctx)
	if err != nil || len(entries) != 1 || entries[r.ID].Recording.Title != review.Title || entries[r.ID].Recording.ConcertID != review.ConcertID {
		t.Fatalf("refresh lost editorial state: %v %v", entries, err)
	}
	s := music.Schedule{Revision: "persistent", Programs: []music.Program{{RecordingID: r.ID, Start: time.Now().UTC(), Duration: 10}}}
	if err = d.SaveSchedule(ctx, s); err != nil {
		t.Fatal(err)
	}
	saved, err := d.Schedule(ctx)
	if err != nil || saved.Revision != s.Revision {
		t.Fatal(saved, err)
	}
	if err = d.Review(ctx, r.ID, music.Review{Kind: "soundboard", Acquisition: true}); err == nil {
		t.Fatal("soundboard acquired")
	}
	entries, err = d.Entries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	e := entries[r.ID]
	approval := PublicApproval{Approved: true, SourceHash: e.SourceHash, MediaSHA256: e.SHA256, PolicyBasis: "test owner decision, noncommercial audience policy", Reviewer: "test"}
	bad := approval
	bad.SourceHash = music.Digest([]byte("wrong"))
	if err = d.ApprovePublic(ctx, r.ID, bad); err == nil {
		t.Fatal("approved stale source")
	}
	if err = d.ApprovePublic(ctx, r.ID, approval); err != nil {
		t.Fatal(err)
	}
	public := &DB{Pool: pool, Public: true}
	if visible, err := public.Entries(ctx); err != nil || len(visible) != 1 {
		t.Fatal("approved recording missing", visible, err)
	}
	if err = public.ApprovePublic(ctx, r.ID, approval); err == nil {
		t.Fatal("serving connection changed approval")
	}
	if err = d.UpsertItems(ctx, []music.SourceItem{{ID: r.ID, ShowDate: "1977-05-08", Title: "Approved", Kind: "audience"}, {ID: "hidden", ShowDate: "1969-06-21", Title: "Private"}}); err != nil {
		t.Fatal(err)
	}
	if years, err := public.Years(ctx); err != nil || len(years) != 1 || years[0].Year != 1977 {
		t.Fatal("private inventory leaked", years, err)
	}
	if hits, n, err := public.Search(ctx, "", "", 1, 20); err != nil || n != 1 || len(hits) != 1 {
		t.Fatal("public search scope", hits, n, err)
	}
	if err = d.Import(ctx, r, []byte(`{"raw":3}`)); err != nil {
		t.Fatal(err)
	}
	if visible, err := public.Entries(ctx); err != nil || len(visible) != 0 {
		t.Fatal("source refresh retained approval", visible, err)
	}
	e2, _ := d.Entries(ctx)
	approval.SourceHash = e2[r.ID].SourceHash
	if err = d.ApprovePublic(ctx, r.ID, approval); err != nil {
		t.Fatal(err)
	}
	if err = d.Media(ctx, r.ID, "replacement.wav", music.Digest([]byte("replacement")), nil); err != nil {
		t.Fatal(err)
	}
	if visible, err := public.Entries(ctx); err != nil || len(visible) != 0 {
		t.Fatal("media replacement retained approval", visible, err)
	}
	approval.Approved = false
	if err = d.ApprovePublic(ctx, r.ID, approval); err != nil {
		t.Fatal(err)
	}
	var history int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM public_approval_history").Scan(&history); err != nil || history != 3 {
		t.Fatal("decision history missing", history, err)
	}

}
