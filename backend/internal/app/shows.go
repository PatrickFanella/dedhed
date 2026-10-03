package app

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/catalog"
	"git.subcult.tv/subculture-collective/dedhed/internal/music"
)

// Show-index endpoints read the collection-wide inventory. They describe what
// exists at the source; only entries that also pass eligible() can be played.

func (s *Server) calendar(w http.ResponseWriter, r *http.Request) {
	years, err := s.DB.Years(r.Context())
	if err != nil {
		slog.Error("calendar read failed", "error", err)
		write(w, 503, map[string]string{"error": "The show index isn't answering. Try again in a moment."})
		return
	}
	out := map[string]any{"years": years}
	if y := r.URL.Query().Get("year"); y != "" {
		year, err := strconv.Atoi(y)
		if err != nil {
			write(w, 400, map[string]string{"error": "That year isn't on the calendar."})
			return
		}
		days, err := s.DB.Days(r.Context(), year)
		if err != nil {
			write(w, 400, map[string]string{"error": "That year isn't on the calendar."})
			return
		}
		out["year"] = year
		out["days"] = days
	}
	write(w, 200, out)
}

// DateItem is one source item on a date, annotated with its library state.
type DateItem struct {
	music.SourceItem
	Playable bool   `json:"playable"`
	Library  string `json:"library"` // "", "imported", or "playable"
}

func (s *Server) date(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	items, err := s.DB.ItemsOn(r.Context(), date)
	if err != nil {
		write(w, 400, map[string]string{"error": "That isn't a date the calendar can read."})
		return
	}
	entries, ok := s.entries(w, r)
	if !ok {
		return
	}
	out := []DateItem{}
	playable := 0
	for _, it := range items {
		d := DateItem{SourceItem: it}
		if e, ok := entries[it.ID]; ok {
			d.Library = "imported"
			if eligible(e) {
				d.Library = "playable"
				d.Playable = true
				playable++
			}
		}
		out = append(out, d)
	}
	// Playable recordings first; the inventory order (rating, downloads) is kept within each group.
	sorted := make([]DateItem, 0, len(out))
	for _, d := range out {
		if d.Playable {
			sorted = append(sorted, d)
		}
	}
	for _, d := range out {
		if !d.Playable {
			sorted = append(sorted, d)
		}
	}
	write(w, 200, map[string]any{"date": date, "items": sorted, "playable": playable})
}

func (s *Server) onThisDay(w http.ResponseWriter, r *http.Request) {
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	t := now().UTC()
	month, day := int(t.Month()), t.Day()
	if m := r.URL.Query().Get("month"); m != "" {
		month, _ = strconv.Atoi(m)
	}
	if d := r.URL.Query().Get("day"); d != "" {
		day, _ = strconv.Atoi(d)
	}
	days, err := s.DB.OnThisDay(r.Context(), month, day)
	if err != nil {
		write(w, 400, map[string]string{"error": "invalid month or day"})
		return
	}
	write(w, 200, map[string]any{"month": month, "day": day, "dates": days})
}

func (s *Server) shows(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	hits, total, err := s.DB.Search(r.Context(), r.URL.Query().Get("q"), r.URL.Query().Get("year"), page, 20)
	if err != nil {
		write(w, 400, map[string]string{"error": "invalid search"})
		return
	}
	if hits == nil {
		hits = []catalog.SearchHit{}
	}
	write(w, 200, map[string]any{"dates": hits, "total": total, "page": page})
}
