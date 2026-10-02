// Package source normalizes audio metadata using the pinned Arkhiv HTTP client.
package source

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"git.subcult.tv/subculture-collective/dedhed/internal/music"
	ia "git.subcult.tv/subculture-collective/dedhed/third_party/arkhiv/internetarchive"
)

type Adapter struct{ Client *ia.Client }

func New() *Adapter { return &Adapter{Client: ia.New()} }
func (a *Adapter) Fetch(ctx context.Context, id string) (music.Recording, []byte, error) {
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,200}$`).MatchString(id) {
		return music.Recording{}, nil, fmt.Errorf("invalid source id")
	}
	_, raw, err := a.Client.Metadata(ctx, id)
	if err != nil {
		return music.Recording{}, raw, err
	}
	r, err := Normalize(id, raw)
	return r, raw, err
}
func text(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []any:
		parts := []string{}
		for _, v := range x {
			parts = append(parts, text(v))
		}
		return strings.Join(parts, "; ")
	default:
		return ""
	}
}
func Duration(v string) (float64, error) {
	parts := strings.Split(v, ":")
	if len(parts) > 3 {
		return 0, fmt.Errorf("invalid duration")
	}
	n := 0.0
	for i, p := range parts {
		x, e := strconv.ParseFloat(p, 64)
		if e != nil || x < 0 || math.IsNaN(x) || math.IsInf(x, 0) || (len(parts) > 1 && i > 0 && x >= 60) {
			return 0, fmt.Errorf("invalid duration %q", v)
		}
		n = n*60 + x
	}
	if n <= 0 || n > 86400 {
		return 0, fmt.Errorf("duration out of bounds")
	}
	return n, nil
}

var digits = regexp.MustCompile(`\d+|\D+`)

func NaturalLess(a, b string) bool {
	x, y := digits.FindAllString(a, -1), digits.FindAllString(b, -1)
	for i := 0; i < len(x) && i < len(y); i++ {
		if x[i] == y[i] {
			continue
		}
		nx, ex := strconv.Atoi(x[i])
		ny, ey := strconv.Atoi(y[i])
		if ex == nil && ey == nil && nx != ny {
			return nx < ny
		}
		return x[i] < y[i]
	}
	return len(x) < len(y)
}
func Normalize(id string, raw []byte) (music.Recording, error) {
	var data struct {
		Metadata map[string]any   `json:"metadata"`
		Files    []map[string]any `json:"files"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return music.Recording{}, err
	}
	m := data.Metadata
	if m == nil {
		return music.Recording{}, fmt.Errorf("missing metadata")
	}
	r := music.Recording{ID: id, Artist: text(m["creator"]), Date: text(m["date"]), Venue: text(m["venue"]), Location: text(m["coverage"]), Title: text(m["title"]), Source: text(m["source"]), Taper: text(m["taper"]), Transferer: text(m["transferer"]), Kind: "unknown", Warnings: []string{}, Tracks: []music.Track{}, Sets: []music.Set{}}
	r.ConcertID = music.Identity(r.Artist, r.Date, r.Venue, r.Location)
	classification := strings.ToLower(r.Source)
	classification = regexp.MustCompile(`behind\s+(the\s+)?soundboard`).ReplaceAllString(classification, "behind mixing position")
	if strings.Contains(classification, "matrix") || strings.Contains(classification, "mtx") || (strings.Contains(classification, "sbd") && strings.Contains(classification, "aud")) {
		r.Kind = "matrix"
	} else if strings.Contains(classification, "soundboard") || strings.Contains(classification, "sbd") {
		r.Kind = "soundboard"
	} else if strings.Contains(classification, "aud") {
		r.Kind = "audience"
	}
	groups := map[string][]map[string]any{}
	for _, f := range data.Files {
		name := text(f["name"])
		format := text(f["format"])
		if format != "VBR MP3" && format != "Flac" && format != "Ogg Vorbis" {
			continue
		}
		if text(f["private"]) == "true" {
			continue
		}
		key := text(f["original"])
		if key == "" {
			key = name
		}
		key = strings.TrimSuffix(key, path.Ext(key))
		groups[key] = append(groups[key], f)
	}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return NaturalLess(keys[i], keys[j]) })
	for _, k := range keys {
		var selected map[string]any
		variants := []music.File{}
		for _, f := range groups[k] {
			size, _ := strconv.ParseInt(text(f["size"]), 10, 64)
			variants = append(variants, music.File{Name: text(f["name"]), Format: text(f["format"]), Bytes: size, MD5: text(f["md5"])})
			if text(f["format"]) == "VBR MP3" {
				selected = f
			}
		}
		if selected == nil {
			r.Warnings = append(r.Warnings, "No MP3 encoding for "+k)
			continue
		}
		dur, err := Duration(text(selected["length"]))
		if err != nil {
			return r, fmt.Errorf("%s: %w", k, err)
		}
		name := text(selected["name"])
		size, _ := strconv.ParseInt(text(selected["size"]), 10, 64)
		title := text(selected["title"])
		if title == "" {
			title = k
			r.Warnings = append(r.Warnings, "Missing track title: "+k)
		}
		r.Tracks = append(r.Tracks, music.Track{Position: len(r.Tracks), Title: title, Start: r.Duration, Duration: dur, File: music.File{Name: name, Format: "VBR MP3", Bytes: size, MD5: text(selected["md5"])}, Variants: variants})
		r.Duration += dur
	}
	if r.Venue == "" {
		r.Warnings = append(r.Warnings, "Venue missing")
	}
	if r.Kind == "unknown" {
		r.Warnings = append(r.Warnings, "Recording classification requires review")
	}
	r.Warnings = append(r.Warnings, "Set boundaries require editorial review")
	return r, r.Validate()
}
