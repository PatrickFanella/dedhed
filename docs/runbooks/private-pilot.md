# Private pilot operations

The pilot is local/private and has no public ingress or relay. Operator commands
require the exact private database/media configuration. Development fixtures,
pilot state, integration tests, and restore destinations are separate.

## Discover and review

From `backend/`, use `go run ./cmd/dedhed discover --ids ITEM1,ITEM2 --out PATH`.
The command permits at most twelve explicit items, fetches only metadata,
preserves raw snapshots, and reports recording classifications and uncertainties.
Collection queries are bounded source-selection work before choosing those IDs.
MP3 derivatives are grouped with their original audio track; rounded durations
are provisional until decoding. Search results are not an acquisition manifest.

Review the individual item, source lineage, credits, track order, missing content,
and policy. Save a review JSON with `kind`, `acquisition`, `playback`, `relay`,
`private_only`, `published`, `notes`, and optional reviewed `sets`.
Set ranges use zero-based inclusive track indices. Keep `published` false until
media is verified. Unknown, matrix, and soundboard sources cannot be acquired
under the pilot policy. Relay is disabled. The audience policy supports source
downloads, not an assumed public rebroadcast license.

## Acquire and prepare audio

Run `python3 scripts/acquire.py --recording NORMALIZED_JSON --review REVIEW_JSON
--root PRIVATE_DATA_ROOT` outside the isolated preview/test networks. The
recording manifest selects exact files; the review must explicitly allow private
audience acquisition. The command permits 50 tracks, 250 MiB per source file,
1 GiB total source bytes, and one transfer at a time.

Transfers enforce byte limits, verify source MD5, compute SHA-256, and retain
originals under hash names. Range resume is accepted only with matching source
checksums and the expected Content-Range. A checksum failure discards the
unverified partial file. Rerunning reuses verified originals. Retries are bounded;
429/5xx honor numeric Retry-After up to 60 seconds. Redirects permit HTTPS
archive.org, its US media hosts, and the observed `dn<number>.ca.archive.org`
delivery hosts; other redirects are rejected.

The script decodes tracks to uniform stereo PCM, measures durations from sample
counts, and creates one continuous AAC/M4A derivative without fades. Temporary
PCM files are removed after encoding. The receipt records source hashes, exact
bytes, decoded durations, ffmpeg version, and derivative probe/hash. Human
audition and browser delivery still require separate evidence.

Copy raw snapshots, normalized metadata, reviewed manifests, and bounded-query
receipts into the private data root's `sources/` directory before backup.

## Import, publish privately, and schedule

From `backend/`, configured for the private DB and media root:

```sh
go run ./cmd/dedhed import --id ITEM --file RAW_METADATA_JSON
go run ./cmd/dedhed review --id ITEM --file REVIEW_JSON
go run ./cmd/dedhed media --id ITEM --media HASH.m4a --sha256 HASH --file PLAYBACK_JSON
```

The media command verifies the local derivative hash. Update the review to
`published: true` only after private playback is approved and verified media
exists. Run the review command again. Publication here means visibility through
the private catalog API; it does not open a public site. Original source metadata
and measured playback metadata are retained separately.

Create a schedule JSON containing a nonempty `revision` and ordered `programs`:
each has `recording_id`, absolute UTC `start`, and positive `duration` in seconds.
Programs must not overlap or exceed the recording's duration. Import with
`go run ./cmd/dedhed schedule --file SCHEDULE_JSON`. Only reviewed published
recordings with media are eligible. A missing source or exhausted schedule
returns an explicit unavailable state; the timeline does not silently change.
Listeners can retry or choose another available set. Reconnect/Return to live
joins server time. Listener controls never write schedule state.

`go run ./cmd/dedhed report` emits catalog state for integrity comparisons.
Verify stored media hashes against that report and originals against acquisition
receipts. Do not treat `/api/health` as evidence that music or schedules work.

## Backup and independent restore

The pilot DB is an explicitly named `dedhed-private-pilot-db` PostgreSQL 17
container. The recovery script permits only Dedhed private/restore containers.
Back up to a new destination with:

```sh
python3 scripts/recovery.py backup --root PRIVATE_DATA_ROOT --backup NEW_BACKUP --db-container dedhed-private-pilot-db
python3 scripts/recovery.py audit --backup NEW_BACKUP
```

The dump uses the matching PostgreSQL client inside the container. The backup
contains source snapshots/reviews, originals, derivatives, acquisition receipts,
catalog, and station schedule with file hashes. Configuration is captured in
the qualification report by nonsecret variable names and effective paths; keep
actual credentials in the existing secret store. Do not back up populated `.env`
files through this script.

Copy the backup to an independent host/destination and verify its hashes there.
A second directory on the same disk is not independent recovery. Create an empty
PostgreSQL 17 DB in a separate `dedhed-restore-*` container. Restore to new storage:

```sh
python3 scripts/recovery.py restore --root NEW_RESTORE_ROOT --backup NEW_BACKUP --db-container dedhed-restore-DATE
```

Restore refuses a populated database or existing storage directory. Compare
record counts, editorial decisions, source/asset hashes, and schedule revision.
Start a separate read-only API and prove on-demand playback and now-playing from
the restored state. Do not replace the active pilot to test recovery.

## Qualification and rollback

Record source revision, effective host/config/storage/URL, exact recordings,
source cuts, playback duration, buffering/errors, mode switching, listener timing,
and resource use. Test known transitions, independent on-demand/radio listeners,
missing-source behavior, restart, and restore. Keep local tests, hosted CI,
browser checks, real audio audition, sustained playback, and recovery separate.

Rollback stops only the task-owned API/frontend processes and retains the private
DB/media/backup. Restore into a new destination before switching consumers.
Public release is tracked by issue #15 and is outside private qualification.

## Sustained delivery observation

A real-time HTTP decode checks sustained server delivery separately from browser
playback. Use a new progress/error/output path under private evidence storage:

```sh
ffmpeg -hide_banner -v warning -re -i http://127.0.0.1:API_PORT/api/audio/ITEM -stats_period 10 -progress PROGRESS_FILE -f null - 2> ERROR_FILE
python3 scripts/observe-pilot.py --api http://127.0.0.1:API_PORT --progress PROGRESS_FILE --errors ERROR_FILE --expected-duration MEASURED_SECONDS --output NEW_REPORT_JSON
```

Run the observer while the decoder runs. It records station/readiness responses
and decoder progress every 15 seconds. Its success requires completed duration,
no decoder warnings/errors, and no observed API failures. It does not prove
speaker output, browser buffering, listener synchronization, or human fidelity.
The observer rejects external API targets and existing report destinations.

Browser qualification uses T3's collaborative browser, with screenshots and
measured audio state recorded separately. This repository has no automated
browser suite; do not report the unit/build gate as browser coverage.
