package app

import (
	"context"
	"io"
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

func TestPublishedAudioRangeAndSourceOutage(t *testing.T) {
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
	schema := "test_" + music.Identity(time.Now().String())
	if _, err = admin.Pool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Pool.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
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
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	s := &Server{DB: db, Root: root}
	server := httptest.NewServer(s.Handler())
	defer server.Close()
	for _, tc := range []struct {
		path   string
		status int
	}{{"/api/ready", 200}, {"/api/recordings", 200}, {"/api/recordings?q=absent", 200}, {"/api/recordings/missing", 404}, {"/api/audio/missing", 404}, {"/api/station", 200}} {
		resp, err := http.Get(server.URL + tc.path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != tc.status {
			t.Fatalf("%s %d", tc.path, resp.StatusCode)
		}
	}
	req, _ := http.NewRequest("GET", server.URL+"/api/audio/synthetic-1977", nil)
	req.Header.Set("Range", "bytes=0-43")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 206 || len(b) != 44 || string(b[:4]) != "RIFF" {
		t.Fatalf("range response %d %d", resp.StatusCode, len(b))
	}
	entries, err := db.Entries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if err = root.Remove(e.Media); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{"/api/audio/synthetic-1977", "/api/station"} {
		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 503 {
			t.Fatal("missing audio did not surface", path, resp.StatusCode)
		}
	}
}
