# Archive reuse and audio contracts

Decision date: 2026-10-02. Tracks issue #2.

Dedhed uses Go/React/PostgreSQL and a narrow, pinned source package from Arkhiv.
The metadata/HTTP client and its tests are vendored byte-for-byte under
`backend/third_party/arkhiv/internetarchive`. `provenance.json` records the source
revision, paths, and SHA-256 hashes. The package keeps GPL-3.0-only ownership;
Dedhed uses the same source license. Audio remains governed by its source terms.

This versioned source boundary avoids importing an internal Go package or
requiring private Git credentials in builds. Updates must replace the two files
from a reviewed Arkhiv revision, refresh provenance, and run their tests plus
Dedhed adapter tests. The dependency direction is Dedhed → pinned source client.
No Arkhiv schema, PDF reader, application, or migration is copied, and its active
checkout remains untouched. A future public Go module can replace this narrow
vendored package without changing Dedhed's audio contracts.

Arkhiv's document acquisition, catalog, and extraction APIs are document-specific.
Dedhed follows its provenance, bounded acquisition, immutable hash storage, and
independent-backup patterns through audio-specific implementations. No companion
Arkhiv change is required for this slice.

A concert has a reviewed identity comprising artist, date, venue, and source
grouping; a date alone is insufficient. Each recording retains an IA item ID,
raw snapshot, lineage, credits, classification, and independent playback,
acquisition, and relay decisions. Tracks group alternate encodings of the same
source audio. Reviewed sets reference ordered track ranges; uncertain boundaries
remain explicitly unknown. Disc splits never imply a set boundary.

Catalog import updates source records while preserving editorial state. Public
API publication is operator-controlled. The private pilot permits selected
audience recordings only; unknown, matrix, and soundboard sources are ineligible
for local acquisition by default.

The proposed local transport is one continuous AAC/M4A derivative per recording,
assembled from verified, preserved source tracks without fades. Decoded source
durations define track and set offsets; generated media must be probed and
checked before publication. One browser audio element serves both modes, supports
HTTP Range seeking, and never changes the station schedule. Transport acceptance
requires actual browser evidence and an audition/continuity report.

Schedules consist of persisted absolute UTC program entries. Now-playing derives
recording/track/offset from server time. Listener controls cannot mutate schedules.
A missing scheduled recording returns an explicit unavailable state while the
timeline continues; the listener may retry or choose an available on-demand set.
There is no silent substitution or unbounded client retry. An exhausted schedule
is unavailable until the operator extends it.
