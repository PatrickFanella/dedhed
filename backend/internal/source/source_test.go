package source

import (
	"fmt"
	"testing"
)

func TestDurations(t *testing.T) {
	for _, s := range []string{"322.89", "05:22.89", "0:05:22.89"} {
		n, err := Duration(s)
		if err != nil || n != 322.89 {
			t.Fatalf("%s %f %v", s, n, err)
		}
	}
	for _, s := range []string{"NaN", "Inf", "-1", "1:70", "", "0"} {
		if _, err := Duration(s); err == nil {
			t.Fatal(s)
		}
	}
}
func TestDerivativesAndNaturalOrder(t *testing.T) {
	raw := []byte(`{"metadata":{"creator":"Grateful Dead","date":"1977-05-08","venue":"Barton Hall","source":"audience"},"files":[{"name":"t10.mp3","original":"t10.flac","format":"VBR MP3","length":"10","size":"100"},{"name":"t2.ogg","original":"t2.flac","format":"Ogg Vorbis","length":"5"},{"name":"t2.mp3","original":"t2.flac","format":"VBR MP3","title":"Two","length":"00:05","size":"100"}]}`)
	r, err := Normalize("test", raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Tracks) != 2 || r.Tracks[0].Title != "Two" || len(r.Tracks[0].Variants) != 2 || r.Duration != 15 || r.Kind != "audience" {
		t.Fatalf("%+v", r)
	}
	for _, kind := range []string{"sbd", "aud + sbd", "matrix", "unknown"} {
		r, err := Normalize("test", []byte(fmt.Sprintf(`{"metadata":{"source":%q},"files":[{"name":"x.mp3","format":"VBR MP3","length":"5"}]}`, kind)))
		if err != nil {
			t.Fatal(err)
		}
		if kind == "aud + sbd" && r.Kind != "matrix" {
			t.Fatal(r.Kind)
		}
	}
	if _, err := Normalize("empty", []byte(`{"metadata":{},"files":[]}`)); err == nil {
		t.Fatal("empty item accepted")
	}
}

func TestAudienceBehindSoundboardIsNotSoundboard(t *testing.T) {
	r, err := Normalize("test", []byte(`{"metadata":{"source":"Master Audience microphones Behind Soundboard"},"files":[{"name":"x.mp3","format":"VBR MP3","length":"5"}]}`))
	if err != nil || r.Kind != "audience" {
		t.Fatal(r.Kind, err)
	}
}
