# First milestone: Grateful Dead concert library and radio

Status: proposal; no audio playback or station deployment implemented.
Observation date: 2026-10-02.

Delivery tracking: [roadmap](../../ROADMAP.md), with linked implementation issues
and milestone exit gates. This document preserves the design and source evidence;
Gitea issues track current task state and acceptance results.

## Listener experience

Start with a small reviewed concert library and two listening modes. Listeners
can browse concerts and play a chosen set or full show from the beginning, with
pause, seek, and track selection. They can also click Tune in to join the shared
station at its current concert position. The stream is one part of the product;
on-demand set playback is also required in the first version.

Show the song, concert date, venue, recording type, taper/transfer credits, and
source link in both modes, plus the next scheduled concert in radio mode.
Preserve set order, audience
ambience, and song transitions. Do not shuffle across a segue or apply music
crossfades over concert transitions.

Both shared radio and on-demand set playback were confirmed by the user on
2026-10-02. Radio listeners follow one station timeline; on-demand listeners
control their own playback without changing that timeline. Switching modes
replaces the current playback rather than playing two audio sources at once.
Custom personal queues, requests, accounts, favorites, other bands, and DJ commentary
can follow after one dependable concert playback path.

## Current Arkhiv evidence

Inspected local source: `/home/onnwee/Work/Subcult/arkhiv-politik`, remote
`https://git.subcult.tv/subculture-collective/arkhiv.git`, HEAD `5ad2573`.
There were uncommitted changes to ingestion and its tests during inspection;
these were left untouched. This is local source evidence, not hosted CI or
runtime qualification.

Arkhiv already contains an Internet Archive metadata client with retries,
bounded acquisition, checksum-addressed storage, raw metadata snapshots,
editorial review/publication state, a PostgreSQL catalog, and recovery tools.
Those mechanisms are useful foundations, subject to audio-specific review.

Its current configuration selects files from a single item, and its catalog,
extraction, publication checks, and reader are document-oriented. Audio/video is
explicitly deferred in its current milestone. Dedhed needs concert entities,
multi-item discovery, track ordering/duration, playback, and scheduling.
PDF extraction and the PDF reader are not radio components.

Prefer extracting reusable source, storage, and provenance packages through an
explicit module boundary once the audio path clarifies their contracts. Keep
Dedhed's station behavior here. Avoid an unmaintained copy of the entire Arkhiv
application or forcing audio into document/variant records.

## Source observations

Music source: [Internet Archive Grateful Dead collection](https://archive.org/details/GratefulDead).
This proposal assumes the music collection, rather than UCSC's memorabilia
archive, is the intended source.

A bounded live metadata query for `collection:GratefulDead AND date:1977-05-08`
returned 23 item matches. Three identifiers were sampled, and metadata was fetched
for one audience recording:
[`gd1977-05-08.ecm33p.moore.berger.miller.117026.flac16`](https://archive.org/metadata/gd1977-05-08.ecm33p.moore.berger.miller.117026.flac16).
This is evidence of multiple source items for one concert, not 23 distinct shows
or proof that all returned recordings are playable.

That item identifies Barton Hall, Ithaca, NY, taper Jerry Moore, and transferers
Charlie Miller and Rob Berger. File metadata includes original FLAC tracks and
MP3/Ogg derivatives. Duration appears as both decimal seconds and `mm:ss`, and
track metadata is not present on every derivative. Preserve raw values, normalize
durations explicitly, and associate derivatives with their original track.
No audio was downloaded or auditioned during this discovery.

The [Archive's Grateful Dead policy](https://archivesupport.zendesk.com/hc/en-us/articles/360004715891-The-Grateful-Dead-Collection)
distinguishes downloadable audience recordings from available soundboards that
are streaming only. It also describes unavailable commercially released shows.
Source availability is not evidence of permission for our public rebroadcast,
mirroring, or monetization. Record those decisions independently. Matrix and
uncertain recordings require review rather than inheriting audience eligibility.

## Playback and source decisions

The shared schedule is confirmed; audio delivery remains open. A shared schedule
can drive client playback from permitted upstream audio
or a continuous station stream; it does not itself require an audio relay.

First prove private playback using a tiny reviewed audience-recording sample and
a clearly labeled synthetic audio fixture. Use metadata discovery separately
from acquisition; never run the document downloader against every listed audio
file. Resolve allowed formats, exact files, byte caps, and reuse conditions in a
reviewed manifest before any acquisition.

For direct upstream playback, verify availability, seeking, cross-origin behavior,
buffering, and actual transitions. This mode depends on Internet Archive uptime.
If it cannot deliver the desired concert continuity, evaluate a continuous stream
with an audio engine and Icecast-compatible delivery, using recordings approved
for that use. A local station stream adds storage, encoding, and distribution
responsibilities. Do not treat that as already covered by Archive access rules.

Soundboards remain outside local acquisition and station relay in the first
milestone. Keep their source metadata and links if useful; introduce playback
only after establishing a permitted access path.

## Data and scheduling

Model the following separately:

- Concert: artist, date, venue, location, and reviewed identity. Date alone is not
  a universal concert identifier.
- Recording: source item, audience/soundboard/matrix/unknown classification,
  lineage, credits, completeness, access state, and editorial selection.
- Set: an ordered group of tracks within a recording, with a reviewed label
  such as Set I, Set II, or Encore. Preserve uncertain boundaries explicitly;
  disc splits are not necessarily set boundaries. Each identified set can be
  started independently and stops at its end by default.
- Track: recording, set/disc/track position, title, duration, segue relationship,
  and alternative encoded files. Lexical filename order alone is insufficient.
- Asset or external reference: source filename/URL, format, checksums when
  supplied, availability, and independent acquisition/playback/relay decisions.
- Station schedule: recording selection, absolute start time, known playable
  duration, revision, and persisted history.

Use a persisted schedule and server time to determine current recording, track,
and offset. New listeners join the same program point; reconnecting listeners
can return to live. If pause is available, show that playback is behind the
station and provide a Return to live action. Browsers require a listener gesture
to begin playback. Synchronization tolerance must be measured; direct client
playback should not promise sample-accurate agreement.

Keep on-demand playback position separate from the station schedule. Starting,
pausing, seeking, or changing a set never mutates the shared program. Tune in
joins the station's current position, including after listening to a set. Use
one player with an explicit radio/on-demand mode and mode-specific controls.

Choose one reviewed recording per concert for the initial schedule. Do not rank
solely by ratings or download counts. Check completeness, cuts, pitch, track
boundaries, source credits, and availability. Keep alternate recordings in the
catalog without scheduling the same concert repeatedly by accident.

## Ordered implementation

1. Build bounded metadata discovery and normalization for a proposed sample of
   10 concerts across several eras, with alternate recordings for at least one
   concert. Produce an operator-readable report and raw snapshots. Record
   missing fields and uncertain classifications instead of guessing.
2. Implement concert browsing, a recording detail view, and on-demand playback
   for complete shows and individual sets. Start with synthetic fixtures, then
   a reviewed source sample. Verify actual audio, seeking, track selection, set
   boundaries, duration, and transitions before choosing the delivery method.
3. Add a persisted station schedule and now-playing endpoint returning server
   time, schedule revision, recording, track, offset, and next program. Prove
   restart, boundary, late join, and reconnect behavior.
4. Add Tune in, Return to live, and upcoming concerts to the listening interface.
   Make radio and on-demand modes clear, with shared volume and recording credits.
   Verify switching between a selected set and live radio. Provide visible
   loading, unavailable-source, and retry states on desktop and mobile.
5. Run a private library and station pilot. Measure continuity and multi-listener timing;
   verify recovery and resource use. Public deployment follows a separate
   source/rebroadcast decision and an explicitly selected host/domain.

## Acceptance evidence

| Requirement | Evidence |
| --- | --- |
| Accurate concert inventory | Reviewed manifest; separate concert, recording, and track counts |
| Repeatable import | Same snapshots produce stable identities; source refresh retains editorial choices |
| Correct source handling | Audience, soundboard, matrix, missing-access, and withdrawn-item checks |
| Concert continuity | Auditioned segue and set boundary; measure any audible gap |
| On-demand sets | Start a selected set from its first track; pause, seek, select a track, and stop at its reviewed end |
| Independent listening | One listener plays a chosen set while another stays on radio; the schedule does not change |
| Mode switching | Tune in from on-demand joins the current station offset with no overlapping audio |
| Shared program | Two browser sessions join mid-track within an agreed timing tolerance |
| Recovery | Restart schedule service mid-show; reconnect and Return to live reach expected position |
| Failure handling | Missing track and source outage exercise a bounded retry and defined fallback |
| Usable listening page | Rendered desktop/mobile and keyboard checks, including playback failures |
| Private pilot | At least one complete concert plays through; capture interruption and resource evidence |
| Public readiness | Reviewed distribution decision, exact runtime identity, and live destination checks |

Fixture tests, source inspection, browser playback, private pilot, hosted CI, and
public deployment are separate proof levels. This document establishes none of
the playback or deployment evidence yet.

## Decisions still open

Choose direct upstream playback versus a continuous stream based on the audio
proof and permitted use.
Select the first reviewed concerts, station name/domain, intended audience, and
whether the station will have ads or other monetization before public release.

API reference: [Internet Archive metadata read API](https://archive.org/developers/md-read.html).
