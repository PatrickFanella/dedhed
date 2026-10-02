package fixture

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"git.subcult.tv/subculture-collective/dedhed/internal/catalog"
	"git.subcult.tv/subculture-collective/dedhed/internal/music"
)

func Seed(ctx context.Context, d *catalog.DB, root string) error {
	if err := os.MkdirAll(root, 0700); err != nil {
		return err
	}
	ids := []string{"synthetic-1977", "synthetic-1989"}
	for i, id := range ids {
		b := WAV(60, 220+float64(i)*110)
		name := music.Digest(b) + ".wav"
		if err := os.WriteFile(filepath.Join(root, name), b, 0600); err != nil {
			return err
		}
		r := music.Recording{ID: id, ConcertID: "concert-" + id, Artist: "Synthetic fixture", Date: []string{"1977-05-08", "1989-07-07"}[i], Venue: []string{"Barton Hall (synthetic)", "JFK Stadium (synthetic)"}[i], Location: []string{"Ithaca, NY", "Philadelphia, PA"}[i], Title: "Synthetic concert — test tones", Kind: "synthetic", Source: "Generated test tones; not a Grateful Dead recording", Synthetic: true, Duration: 60, Tracks: []music.Track{}, Sets: []music.Set{}, Warnings: []string{}}
		for j := 0; j < 4; j++ {
			r.Tracks = append(r.Tracks, music.Track{Position: j, Title: fmt.Sprintf("Test tone %d", j+1), Start: float64(j) * 15, Duration: 15, Variants: []music.File{}})
		}
		raw, _ := json.Marshal(r)
		if err := d.Import(ctx, r, raw); err != nil {
			return err
		}
		if err := d.Media(ctx, id, name, music.Digest(b), nil); err != nil {
			return err
		}
		if err := d.Review(ctx, id, music.Review{Published: true, Playback: true, PrivateOnly: true, Kind: "synthetic", Notes: "Synthetic development fixture", Sets: []music.Set{{Title: "Set I", First: 0, Last: 1, Reviewed: true}, {Title: "Set II", First: 2, Last: 3, Reviewed: true}}}); err != nil {
			return err
		}
	}
	// Idempotent seed leaves an existing station timeline intact.
	if _, err := d.Schedule(ctx); err == nil {
		return nil
	}
	start := time.Now().UTC()
	s := music.Schedule{Revision: music.Identity(start.String()), Programs: []music.Program{}}
	for i := 0; i < 1440; i++ {
		s.Programs = append(s.Programs, music.Program{RecordingID: ids[i%2], Start: start.Add(time.Duration(i) * time.Minute), Duration: 60})
	}
	return d.SaveSchedule(ctx, s)
}
func WAV(seconds int, hz float64) []byte {
	const rate = 16000
	var out bytes.Buffer
	out.WriteString("RIFF")
	_ = binary.Write(&out, binary.LittleEndian, uint32(36+seconds*rate*2))
	out.WriteString("WAVEfmt ")
	_ = binary.Write(&out, binary.LittleEndian, uint32(16))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint32(rate))
	_ = binary.Write(&out, binary.LittleEndian, uint32(rate*2))
	_ = binary.Write(&out, binary.LittleEndian, uint16(2))
	_ = binary.Write(&out, binary.LittleEndian, uint16(16))
	out.WriteString("data")
	_ = binary.Write(&out, binary.LittleEndian, uint32(seconds*rate*2))
	for i := 0; i < seconds*rate; i++ {
		_ = binary.Write(&out, binary.LittleEndian, int16(2000*math.Sin(2*math.Pi*hz*float64(i)/rate)))
	}
	return out.Bytes()
}
