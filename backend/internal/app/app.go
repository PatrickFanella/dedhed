package app

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/catalog"
	"git.subcult.tv/subculture-collective/dedhed/internal/music"
)

type Server struct {
	DB       *catalog.DB
	Root     *os.Root
	Revision string
	Now      func() time.Time
}

func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		write(w, http.StatusOK, map[string]string{"status": "ok", "revision": s.Revision})
	})
	m.HandleFunc("GET /api/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if s.DB.Pool.Ping(ctx) != nil {
			write(w, 503, map[string]string{"error": "catalog unavailable"})
			return
		}
		write(w, 200, map[string]string{"status": "ready"})
	})
	m.HandleFunc("GET /api/recordings", s.list)
	m.HandleFunc("GET /api/recordings/{id}", s.detail)
	m.HandleFunc("GET /api/audio/{id}", s.audio)
	m.HandleFunc("GET /api/station", s.station)
	m.HandleFunc("GET /api/calendar", s.calendar)
	m.HandleFunc("GET /api/dates/{date}", s.date)
	m.HandleFunc("GET /api/on-this-day", s.onThisDay)
	m.HandleFunc("GET /api/shows", s.shows)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		m.ServeHTTP(w, r)
	})
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func (s *Server) entries(w http.ResponseWriter, r *http.Request) (map[string]music.Entry, bool) {
	e, err := s.DB.Entries(r.Context())
	if err != nil {
		slog.Error("catalog read failed", "error", err)
		write(w, 503, map[string]string{"error": "catalog unavailable"})
		return nil, false
	}
	return e, true
}
func eligible(e music.Entry) bool { return e.Review.Published && e.Review.Playback && e.Media != "" }
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	entries, ok := s.entries(w, r)
	if !ok {
		return
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	year := r.URL.Query().Get("year")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	if page > 10000 {
		write(w, 400, map[string]string{"error": "page out of range"})
		return
	}
	list := []music.Entry{}
	for _, e := range entries {
		m := e.Recording
		if eligible(e) && (q == "" || strings.Contains(strings.ToLower(m.Title+" "+m.Venue+" "+m.Location+" "+m.Date), q)) && (year == "" || strings.HasPrefix(m.Date, year)) {
			list = append(list, e)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Recording.Date != list[j].Recording.Date {
			return list[i].Recording.Date < list[j].Recording.Date
		}
		return list[i].Recording.ID < list[j].Recording.ID
	})
	total := len(list)
	start := (page - 1) * 20
	if start > total {
		start = total
	}
	end := start + 20
	if end > total {
		end = total
	}
	write(w, 200, map[string]any{"items": list[start:end], "total": total, "page": page})
}
func (s *Server) detail(w http.ResponseWriter, r *http.Request) {
	entries, ok := s.entries(w, r)
	if !ok {
		return
	}
	e, ok := entries[r.PathValue("id")]
	if !ok || !eligible(e) {
		write(w, 404, map[string]string{"error": "recording unavailable"})
		return
	}
	write(w, 200, e)
}
func (s *Server) audio(w http.ResponseWriter, r *http.Request) {
	entries, ok := s.entries(w, r)
	if !ok {
		return
	}
	e, ok := entries[r.PathValue("id")]
	if !ok || !eligible(e) {
		write(w, 404, map[string]string{"error": "recording unavailable"})
		return
	}
	f, err := s.Root.Open(e.Media)
	if err != nil {
		write(w, 503, map[string]string{"error": "audio source unavailable; choose another recording or retry"})
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		write(w, 503, map[string]string{"error": "audio source unavailable"})
		return
	}
	w.Header().Set("ETag", `"`+e.SHA256+`"`)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	http.ServeContent(w, r, e.Media, stat.ModTime(), f)
}
func (s *Server) station(w http.ResponseWriter, r *http.Request) {
	entries, ok := s.entries(w, r)
	if !ok {
		return
	}
	schedule, err := s.DB.Schedule(r.Context())
	if err != nil {
		write(w, 503, map[string]string{"error": "station schedule unavailable"})
		return
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	state, err := schedule.At(now().UTC(), entries)
	if err == nil {
		_, err = s.Root.Stat(entries[state.Program.RecordingID].Media)
	}
	if err != nil {
		write(w, 503, map[string]string{"error": "station unavailable; retry or play a set from the library"})
		return
	}
	write(w, 200, state)
}
