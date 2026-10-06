# Audio viewer qualification — 2026-10-06

Dedhed retains its cassette interface, reviewed tracks/sets, native backend and separate core 0.1.2 pin. The audio SDK is independently pinned to 0.3.1 with registry lock and provenance. Native recording IDs remain browser-local favorites; account sync and audio chat require the separately tracked canonical identity/room integration.

## Required verification

Host: sputnik. Source: `/tmp/arkhiv-consumers/dedhed`. Base revision: `1b783e77a98c49b3b9404a4ed6c2e9c579df40ee`, with the viewer changes uncommitted during qualification. The isolated launcher used profile `dedhed` and an explicitly disposable PostgreSQL test service. The final full `make verify` run (session 36582) exited 0: Go vet/race/integration, seven Python checks, five frontend tests, build and pin checks. Hosted CI remains owner-waived during the Kvant outage.

Catalog regression uses 27 generated recordings to verify venue and original-date filtering before pagination, literal matching, invalid dates returning 400 and retention of unknown dates without bounds.

## Native collaborative browser

The owned preview at port 35811 used two generated 60-second recordings, each with four tones and two reviewed sets. Set II started at 30.50 seconds, pause stopped playback, seek reached 45.00, and resume advanced to 45.14. Selecting track 2 reached 15.30 seconds. Set II ended at 60.00 with audio paused and an explicit ended state. Full-recording playback and reviewed-set selection used one audio element. Native Enter on the focused transport paused and resumed it.

Radio had no position scrubber and showed scheduled replay and the original concert date. Pausing said the station continues; resuming fetched the current program. A resume across a generated program boundary joined the new recording within 0.028 seconds of the measured station position. A separate 390px same-origin frame played a selected set while the parent joined radio, then joined the same station. Both were ready and advancing, separated by 0.014 seconds; child drift from the server was 0.106 seconds, below the documented three-second tolerance. The schedule revision stayed `6d2889d59d175a0c675fa448`.

Injecting a missing audio URL produced an actual HTTP 404, an unavailable label and retry control. Retry fetched the current station, restored ready/advancing playback and cleared the error. Source failures were not labelled autoplay denial.

Browser controls stored volume 0.3 and mute across remounts and mode changes. Both controls were visible in the 390px frame; the page width was 384px. This is desktop-browser responsive evidence, not physical-device testing. Tape-rack/table switching retained the cassette details, metadata and source credits. A saved favorite survived reload, an unavailable reference remained visible and removable, and manual share fallback exposed the recording link. Listening focus activated and Escape restored the page.

Synthetic tones qualify control behavior and timing. They do not prove human listening to every real concert segue or complete-corpus audiovisual fidelity. Those acceptance requirements remain tracked in the real-pilot issues. The preview is stopped with development data retained. No acquisition, schedule edit, new public approval or live provider activation occurred.


## Published follow-up adoption

The consumer now independently pins SDK 0.4.0 and core 0.3.0 from Arkhiv source
39a49fb45005b001e14218823c22ed19c3de1177. SDK registry bytes were independently
downloaded and verified against its SHA-256 receipt. The immutable core receipt
updates the compatibility browse filters and optional media-specific chat storage.
The native backend and current deployment remain independently owned; this pin
update does not perform live migration or activate account/chat hosting.

Full `make verify` passed in an isolated T3 snapshot for these exact dependency
locks and core receipt. The print dump/restore checks used an explicitly
disposable PostgreSQL 15 service matching the tools image's pg_dump client; the
earlier PostgreSQL 17 client mismatch was a launcher failure, not a passing test.
Hosted CI remains owner-waived during the Kvant outage.
