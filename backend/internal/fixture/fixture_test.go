package fixture

import (
	"encoding/binary"
	"testing"
)

func TestFixtureTracksHaveDistinctAudio(t *testing.T) {
	const rate = 16000
	wav := WAV(60, 220)
	if len(wav) != 44+60*rate*2 {
		t.Fatal("fixture duration changed")
	}
	for track, expected := range []int{220, 275, 330, 385} {
		// Count rising crossings in the middle second of each 15-second track.
		start := (track*15 + 5) * rate
		crossings := 0
		for i := start + 1; i < start+rate; i++ {
			previous := int16(binary.LittleEndian.Uint16(wav[44+(i-1)*2:]))
			current := int16(binary.LittleEndian.Uint16(wav[44+i*2:]))
			if previous <= 0 && current > 0 {
				crossings++
			}
		}
		if crossings < expected-1 || crossings > expected+1 {
			t.Fatalf("track %d frequency %d, expected %d", track, crossings, expected)
		}
	}
}
