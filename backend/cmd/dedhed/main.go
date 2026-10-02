package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/app"
	"git.subcult.tv/subculture-collective/dedhed/internal/catalog"
	"git.subcult.tv/subculture-collective/dedhed/internal/fixture"
	"git.subcult.tv/subculture-collective/dedhed/internal/inventory"
	"git.subcult.tv/subculture-collective/dedhed/internal/music"
	"git.subcult.tv/subculture-collective/dedhed/internal/source"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: dedhed migrate|seed|discover|inventory-scrape|inventory-fetch|inventory-import|import|review|media|schedule|report|serve")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	cmd := os.Args[1]
	f := flag.NewFlagSet(cmd, flag.ContinueOnError)
	id := f.String("id", "", "recording id")
	file := f.String("file", "", "JSON input")
	ids := f.String("ids", "", "up to twelve explicit IA item ids")
	out := f.String("out", "", "snapshot output directory")
	media := f.String("media", "", "managed media basename")
	sha := f.String("sha256", "", "verified media SHA-256")
	dir := f.String("dir", "", "inventory directory")
	query := f.String("query", "collection:GratefulDead", "source collection query")
	limit := f.Int("limit", 0, "maximum items to fetch this run (0 = all)")
	concurrency := f.Int("concurrency", 2, "parallel metadata requests (1-4)")
	delay := f.Duration("delay", 250*time.Millisecond, "pause after each metadata request")
	if err := f.Parse(os.Args[2:]); err != nil {
		return err
	}
	if cmd == "discover" {
		list := strings.Split(*ids, ",")
		if *ids == "" || len(list) > 12 || *out == "" {
			return fmt.Errorf("discover requires 1–12 ids and an output directory")
		}
		if err := os.MkdirAll(*out, 0700); err != nil {
			return err
		}
		adapter := source.New()
		report := []map[string]any{}
		for _, item := range list {
			r, raw, err := adapter.Fetch(ctx, item)
			if err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(*out, item+".json"), raw, 0600); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(r, "", "  ")
			if err = os.WriteFile(filepath.Join(*out, item+".normalized.json"), b, 0600); err != nil {
				return err
			}
			report = append(report, map[string]any{"item": item, "concert": r.ConcertID, "tracks": len(r.Tracks), "duration": r.Duration, "kind": r.Kind, "warnings": r.Warnings, "snapshot_sha256": music.Digest(raw)})
		}
		return json.NewEncoder(os.Stdout).Encode(report)
	}
	if cmd == "inventory-scrape" || cmd == "inventory-fetch" {
		if *dir == "" {
			return fmt.Errorf("%s requires -dir", cmd)
		}
		if err := os.MkdirAll(*dir, 0700); err != nil {
			return err
		}
		adapter := source.New()
		if cmd == "inventory-scrape" {
			tmp := filepath.Join(*dir, "collection.jsonl.part")
			out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				return err
			}
			written, total, err := inventory.Scrape(ctx, adapter.Client, *query, out)
			out.Close()
			if err != nil {
				return err
			}
			if err = os.Rename(tmp, filepath.Join(*dir, "collection.jsonl")); err != nil {
				return err
			}
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"written": written, "source_total": total, "file": filepath.Join(*dir, "collection.jsonl")})
		}
		summaries, err := inventory.ReadSummaries(filepath.Join(*dir, "collection.jsonl"))
		if err != nil {
			return err
		}
		ids := []string{}
		for _, s := range summaries {
			if _, err := os.Stat(filepath.Join(*dir, "items", s.Identifier+".json")); err != nil {
				ids = append(ids, s.Identifier)
			}
		}
		if *limit > 0 && len(ids) > *limit {
			ids = ids[:*limit]
		}
		started := time.Now()
		p, err := inventory.FetchAll(ctx, adapter, ids, *dir, *concurrency, *delay, func(p inventory.Progress) {
			if (p.Done+p.Failed+p.Skipped)%100 == 0 {
				log.Printf("inventory fetch %d/%d done, %d failed, %s elapsed", p.Done+p.Skipped, p.Total, p.Failed, time.Since(started).Round(time.Second))
			}
		})
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"pending_before": len(ids), "fetched": p.Done, "failed": p.Failed, "elapsed": time.Since(started).String()})
		return err
	}
	d, err := catalog.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("database connection: %w", err)
	}
	defer d.Pool.Close()
	root := env("DEDHED_MEDIA", "../var/media")
	switch cmd {
	case "migrate":
		return d.Migrate(ctx, env("DEDHED_MIGRATIONS", "../migrations"))
	case "seed":
		return fixture.Seed(ctx, d, root)
	case "inventory-import":
		if *dir == "" {
			return fmt.Errorf("inventory-import requires -dir")
		}
		total, fetched, err := inventory.Load(*dir, 200, func(batch []music.SourceItem) error { return d.UpsertItems(ctx, batch) })
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"items": total, "with_metadata": fetched})
	case "import":
		b, err := os.ReadFile(*file)
		if err != nil {
			return err
		}
		r, err := source.Normalize(*id, b)
		if err != nil {
			return err
		}
		return d.Import(ctx, r, b)
	case "review":
		b, err := os.ReadFile(*file)
		if err != nil {
			return err
		}
		var review music.Review
		if err = json.Unmarshal(b, &review); err != nil {
			return err
		}
		return d.Review(ctx, *id, review)
	case "media":
		var recording *music.Recording
		if *file != "" {
			b, err := os.ReadFile(*file)
			if err != nil {
				return err
			}
			recording = &music.Recording{}
			if err = json.Unmarshal(b, recording); err != nil {
				return err
			}
		}
		b, err := os.ReadFile(filepath.Join(root, *media))
		if err != nil {
			return err
		}
		if music.Digest(b) != *sha {
			return fmt.Errorf("media hash mismatch")
		}
		return d.Media(ctx, *id, *media, *sha, recording)
	case "schedule":
		b, err := os.ReadFile(*file)
		if err != nil {
			return err
		}
		var schedule music.Schedule
		if err = json.Unmarshal(b, &schedule); err != nil {
			return err
		}
		return d.SaveSchedule(ctx, schedule)
	case "report":
		entries, err := d.Entries(ctx)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(entries)
	case "serve":
		mediaRoot, err := os.OpenRoot(root)
		if err != nil {
			return err
		}
		defer mediaRoot.Close()
		s := &app.Server{DB: d, Root: mediaRoot, Revision: env("DEDHED_REVISION", "development")}
		httpServer := &http.Server{Addr: env("API_ADDR", "127.0.0.1:8080"), Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
		go func() {
			<-ctx.Done()
			shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = httpServer.Shutdown(shutdown)
		}()
		log.Printf("Dedhed API listening on %s", httpServer.Addr)
		err = httpServer.ListenAndServe()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}
func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
