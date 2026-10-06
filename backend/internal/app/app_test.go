package app

import (
	"context"
	"encoding/json"
	"fmt"
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
	seeded, err := db.Entries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	alternate := seeded["synthetic-1977"]
	alternate.Recording.ID = "synthetic-alternate"
	alternate.Recording.Venue = ""
	alternate.Recording.Sets = []music.Set{}
	if err = db.Import(ctx, alternate.Recording, []byte(`{"alternate":true}`)); err != nil {
		t.Fatal(err)
	}
	if err = db.Media(ctx, alternate.Recording.ID, alternate.Media, alternate.SHA256, nil); err != nil {
		t.Fatal(err)
	}
	alternate.Review.Sets = nil
	if err = db.Review(ctx, alternate.Recording.ID, alternate.Review); err != nil {
		t.Fatal(err)
	}
	hidden := alternate.Recording
	hidden.ID = "unpublished"
	if err = db.Import(ctx, hidden, []byte(`{"hidden":true}`)); err != nil {
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
	for _, query := range []struct {
		path  string
		total int
	}{{"/api/recordings", 3}, {"/api/recordings?q=absent", 0}} {
		response, err := http.Get(server.URL + query.path)
		if err != nil {
			t.Fatal(err)
		}
		var result struct {
			Items []music.Entry `json:"items"`
			Total int           `json:"total"`
		}
		err = json.NewDecoder(response.Body).Decode(&result)
		response.Body.Close()
		if err != nil || result.Total != query.total || len(result.Items) != query.total {
			t.Fatal("publication filter or empty result", result, err)
		}
	}
	for _, tc := range []struct {
		path   string
		status int
	}{{"/api/ready", 200}, {"/api/recordings", 200}, {"/api/recordings?q=absent", 200}, {"/api/recordings/missing", 404}, {"/api/recordings/unpublished", 404}, {"/api/audio/unpublished", 404}, {"/api/audio/missing", 404}, {"/api/station", 200}} {
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
	// A public server must never inherit private publication eligibility.
	pub := seeded["synthetic-1977"]
	rejected := catalog.PublicApproval{Approved: true, SourceHash: pub.SourceHash, MediaSHA256: pub.SHA256, PolicyBasis: "test", Reviewer: "test"}
	if err = db.ApprovePublic(ctx, pub.Recording.ID, rejected); err == nil {
		t.Fatal("synthetic public approval accepted")
	}
	pub.Recording.ID = "audience-public"
	pub.Recording.Synthetic = false
	pub.Recording.Kind = "audience"
	raw := []byte(`{"public":true}`)
	if err = db.Import(ctx, pub.Recording, raw); err != nil {
		t.Fatal(err)
	}
	if err = db.Media(ctx, pub.Recording.ID, pub.Media, pub.SHA256, nil); err != nil {
		t.Fatal(err)
	}
	pub.Review.Kind = "audience"
	if err = db.Review(ctx, pub.Recording.ID, pub.Review); err != nil {
		t.Fatal(err)
	}
	approval := catalog.PublicApproval{Approved: true, SourceHash: music.Digest(raw), MediaSHA256: pub.SHA256, PolicyBasis: "test noncommercial audience", Reviewer: "test"}
	if err = db.ApprovePublic(ctx, pub.Recording.ID, approval); err != nil {
		t.Fatal(err)
	}
	publicDB := &catalog.DB{Pool: db.Pool, Public: true}
	publicServer := httptest.NewServer((&Server{DB: publicDB, Root: root}).Handler())
	defer publicServer.Close()
	for _, tc := range []struct {
		path   string
		status int
	}{{"/api/recordings/audience-public", 200}, {"/api/recordings/synthetic-1977", 404}, {"/api/audio/synthetic-1977", 404}, {"/api/audio/audience-public", 200}} {
		response, err := http.Get(publicServer.URL + tc.path)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != tc.status {
			t.Fatal(tc.path, response.StatusCode)
		}
	}
	approval.Approved = false
	if err = db.ApprovePublic(ctx, pub.Recording.ID, approval); err != nil {
		t.Fatal(err)
	}
	response, err := http.Get(publicServer.URL + "/api/audio/audience-public")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 404 {
		t.Fatal("revoked audio still served", response.StatusCode)
	}
	// More than one page makes accidental page-local filtering observable.
	for i := 0; i < 27; i++ {
		recording := alternate.Recording
		recording.ID = fmt.Sprintf("filter-fixture-%02d", i)
		recording.Venue = "Filter venue"
		recording.Date = "1977-05-08"
		recording.Taper = "Filter credit"
		if i == 26 {
			recording.Date = "1977"
		}
		if err = db.Import(ctx, recording, []byte(`{"synthetic_filter_fixture":true}`)); err != nil {
			t.Fatal(err)
		}
		if err = db.Media(ctx, recording.ID, alternate.Media, alternate.SHA256, nil); err != nil {
			t.Fatal(err)
		}
		if err = db.Review(ctx, recording.ID, alternate.Review); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []struct {
		path                 string
		status, total, count int
	}{
		{"/api/recordings?venue=Filter&q=credit&date_from=1977-01-01&date_to=1977-12-31&page=2", 200, 26, 6},
		{"/api/recordings?venue=Filter&page=2", 200, 27, 7},
		{"/api/recordings?date_from=1977-02-30", 400, 0, 0},
		{"/api/recordings?date_from=1989-01-01&date_to=1977-01-01", 400, 0, 0},
	} {
		response, err := http.Get(server.URL + query.path)
		if err != nil {
			t.Fatal(err)
		}
		var result struct {
			Items []music.Entry `json:"items"`
			Total int           `json:"total"`
		}
		if query.status == 200 {
			err = json.NewDecoder(response.Body).Decode(&result)
		}
		response.Body.Close()
		if response.StatusCode != query.status || err != nil || result.Total != query.total || len(result.Items) != query.count {
			t.Fatal(query.path, response.StatusCode, result, err)
		}
	}
	entries, err := db.Entries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Media == "" {
			continue
		}
		if err = root.Remove(e.Media); err != nil && !os.IsNotExist(err) {
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
