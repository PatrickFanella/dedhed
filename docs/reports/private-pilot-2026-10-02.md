# Private pilot evidence — October 2, 2026

The concert library, on-demand player, and shared schedule are implemented and
available in a loopback-only private pilot. Full listening qualification remains
open. The owner confirmed normal playback but reported poor sound quality for
the initial tape. Mobile controls, two simultaneous listeners, and the complete
browser mode-switching walkthrough have not been verified.

## Runtime and source identity

- Host: Kvant. Checkout: `/home/onnwee/Work/Subcult/dedhed`.
- Private web: `http://127.0.0.1:18577`; API: `127.0.0.1:18578`.
- Runtime backend revision: `935504d864f354e2ea42ed28e6dfc3aa19a4876b`.
  Binary SHA-256: `0623cff1cd82792821ee9bde56e94f2314acccd99067dabacf33f5e721248a12`.
  Later verification and catalog guard commits are not claimed as running code.
- Web revision: `434a49c9a7c47aa895b1adc93e0901c2b3ffb314`; served asset
  `/assets/index-75atJPYu.js` has SHA-256
  `9d10d86281d1fa947f0b8ac647dab94d0cf0388f41ffa3452b26cc69cf192fe0`.
  The served bytes matched the locally built file.
- Primary storage: `~/.local/share/dedhed/pilot`, outside Git. PostgreSQL 17 uses
  the owned `dedhed-private-pilot-db` container/volume on an internal Docker
  network. Effective addresses and nonsecret configuration are in `runtime.json`.
- Arkhiv reuse: the metadata client and its tests, pinned byte-for-byte to
  `2c37f49807beda90c9d9a5aeae796cd291b8695a`. No Arkhiv audio implementation or
  document application was copied. See the [architecture decision](../adr/0001-archive-reuse-and-audio.md).
- Source inventory: 11 recordings across 10 concert identities, with 254 tracks, 642 normalized audio-file variants and 1,443 total metadata-listed files,
  alternate-source metadata, and explicit uncertainties. Metadata discovery is
  separate from the two reviewed private playback selections below.

## Recordings and transport

| Recording | Preserved files/bytes | Measured duration | Private use |
| --- | ---: | ---: | --- |
| `gd1969-06-21.aud-early.cotsman.8995.sbeok.shnf` | 12 MP3s / 75,524,608 | 3,418.3025 seconds | Initial library and radio proof |
| `gd1977-05-08.ecm33p.moore.berger.miller.117026.flac16` | 21 MP3s / 248,501,837 | 10,166.24 seconds | Quality comparison candidate, on demand |

The Fillmore East early-show source credits Jeff Cook and Jeff Cotsman and lists
`AUD>MR>C>R>R>R>D>CD>EAC>SHN` lineage. Its notes disclose a cut in High Time and
splices in Drums and Cryptical. The retained recording is not an uncut concert.
One early-show set was reviewed against the source description.

The Cornell audience source credits Jerry Moore, Rob Berger, and Charlie Miller.
Its notes describe master/first-generation transfers, a same-rig Scarlet/Fire
patch, and an original-master encore. Set I, Set II, and encore boundaries were
reviewed against the explicit source set list and matching track titles. The
recording has not been accepted for subjective sound quality.

Acquisition retained exact files under SHA-256 names, verified Internet Archive
MD5 values, and permitted one transfer at a time. Each file is capped at 250 MiB,
each selection at 1 GiB and 50 tracks. Both completed with no unresolved receipt
errors. Relay is disabled and public distribution is not approved.

Each recording is decoded to uniform PCM and encoded as one continuous stereo
48 kHz AAC/M4A file. The derivative hashes are:

- Fillmore: `8e989e12f66c42a9af64f4eb9c758ff58a08b7dd318d82f527e35076ce62ff2c`.
- Cornell: `7c34d5d102000ac7171f8c6714731fc4757bf6a00b5c2aaff731ba148991e71c`.

HTTP Range delivery supports seeking to tracks and sets. Songs within a recording
share one media source; selecting a different song does not replace the audio
file. Source cuts are retained, and the preparation adds no fades.

## Listening and transition evidence

The native T3 browser rendered the synthetic catalog at desktop width. Set II
started at approximately 30.017 seconds and stopped at 60 seconds in a 60-second
fixture. Radio joined the same fixture using server time. A later fixture improvement gives
its four tracks distinguishable pitches. A complete real-time HTTP decode produced
exactly 960,000 mono samples at 16 kHz, with the expected 220/275/330/385 Hz order
and no added duration. That decoder check is separate from rendered browser proof.

The real Fillmore file started in the native T3 browser with a measured duration
of 3,418.3025 seconds. An attached observer captured approximately 100 seconds of
advancing playback without observed waiting, stalled, error, or ended events.
A separate radio join measured approximately -0.274 seconds relative to the
station offset. That single-client measurement does not establish two-listener
synchronization. The owner reported that playback was working but did not sound
good; the owner later closed the Browser panel.

A 30–60 second source/derivative comparison measured RMS levels of 0.095139 and
0.094872, with peaks below full scale and signal-to-error ratio of 35.13 dB. Each
of the 11 song joins was compared against two seconds from each neighboring
original: all four-second windows contained 384,000 stereo sample values and
remained aligned. Signal-to-error ratios ranged from 29.27 to 36.71 dB. These
checks found no added sample count or silent join interval; they do not replace
human audition or establish subjective fidelity. [Issue #19](https://git.subcult.tv/subculture-collective/dedhed/issues/19)
tracks the listening concern.

A real-time full-file HTTP decode and station/readiness observation is running.
Its final receipt must be checked before claiming sustained completion. This is
decoder/server evidence, separate from browser playback and speaker output.

The T3 automation client changed during verification. Reopening and creating a
new tab still returned a Chrome error document; environment-port navigation and
snapshots failed on client `preview-103871ddea445975277a082ff6188edc`. The owner
confirmed that the pilot had been playing in the T3 Browser panel. Native iOS
simulation is unavailable on this Linux host, and the Android SDK is absent.
No substitute browser result is claimed.

## Schedule and recovery

The persisted schedule revision is `private-pilot-20261002`, beginning at
`2026-10-02T21:11:26.41784Z`. It contains 12 explicit repetitions of the initial
recording for this technical trial. This is a test schedule, not a curated station
program. Adding Cornell to the private library did not alter that schedule.
Restarting only the owned API preserved the schedule and advancing station time.
Two independent API clients returned identical programs with offsets 0.00318 seconds
apart; this checks server consistency, not browser synchronization.

The initial backup contains 59 hashed files and a matching PostgreSQL dump. It was
copied to `nuc:/home/onnwee/backups/dedhed/private-pilot-20261002/` and every hash
was verified on that independent host. A fresh restore to
`~/.local/share/dedhed/restore-20261002` and the separate
`dedhed-restore-20261002` container exactly matched the initial database: ten
recordings, ten snapshots, editorial/media state, and all 12 schedule programs.
The restored API at `127.0.0.1:18580` returned readiness, station state, and the
same 1,024-byte audio Range response with HTTP 206. A seek into the final song also decoded 15 seconds successfully from the restored
API, without decoder diagnostics. This is not a restored browser audition. A newer 90-file comparison backup includes both recordings and the
eleventh metadata snapshot; all 90 files passed their hash audit on NUC. That snapshot was fetched back from
NUC and restored to `~/.local/share/dedhed/restore-comparison-20261002` and a new
`dedhed-restore-comparison-20261002` database. All eleven records/snapshots, both
publication decisions, editorial/media state, and the complete schedule matched
the primary. The separate restored API at `127.0.0.1:18581` used backend revision
`434a49c`; both recordings passed Range requests and 15-second seek/decode checks
without diagnostics.

Recovery does not replace active listeners. Use the [private runbook](../runbooks/private-pilot.md)
for backup, integrity audit, fresh restore, observation, and rollback. Stop only
owned API/frontend processes; retain primary media, databases, and backups.

Foundations milestone M1 is complete. Issues #2–#7 and #10 are closed with
acceptance evidence; browser/listening and dependent recovery gates remain open.

## Verification and open criteria

- The full isolated `t3-dev --profile dedhed verify` gate passed, including real
  disposable PostgreSQL integration, Go race/vet, provenance/link checks, Python
  tests, frontend type/lint/unit checks, and both builds. Test storage is separate
  from the pilot. Source sync and cleanup were verified for the synthetic preview.
- Hosted [run 11984](https://git.subcult.tv/subculture-collective/dedhed/actions/runs/11984)
  passed for CI correction `21f0ffe`; [run 11988](https://git.subcult.tv/subculture-collective/dedhed/actions/runs/11988)
  passed for guard/observer revision `3f0fa4a`. [Run 11991](https://git.subcult.tv/subculture-collective/dedhed/actions/runs/11991)
  passed for report/API regression revision `2cf27e1`. Hosted results are distinct from
  local checks and runtime identity.
- The original CI job could not reach PostgreSQL via job localhost. Its runner
  also lacked service-name DNS on Docker bridge. The fix resolves exactly the
  database labeled for the current run and attempt, with strict cardinality.
- The T3 launcher passed 11 local regression checks. Its [PR #5](https://git.subcult.tv/PatrickFanella/t3-dev-environments/pulls/5)
  merged. The application implementation [PR #17](https://git.subcult.tv/subculture-collective/dedhed/pulls/17)
  merged; [PR #18](https://git.subcult.tv/subculture-collective/dedhed/pulls/18)
  merged the CI fix, and [PR #20](https://git.subcult.tv/subculture-collective/dedhed/pulls/20)
  merged catalog guards and the sustained observer. Hosted [run 11996](https://git.subcult.tv/subculture-collective/dedhed/actions/runs/11996) passed
  for exact code revision `434a49c`.

Issues [#8](https://git.subcult.tv/subculture-collective/dedhed/issues/8),
[#9](https://git.subcult.tv/subculture-collective/dedhed/issues/9),
[#11](https://git.subcult.tv/subculture-collective/dedhed/issues/11),
[#13](https://git.subcult.tv/subculture-collective/dedhed/issues/13), and
[#14](https://git.subcult.tv/subculture-collective/dedhed/issues/14) retain unmet
browser/listening criteria. Those include mobile/keyboard controls, independent
simultaneous set/radio listeners, reconnect and failure UI, auditioned segues,
and a complete sustained browser run. Public release remains
[#15](https://git.subcult.tv/subculture-collective/dedhed/issues/15).

Raw metadata, acquisition and comparison receipts, observer reports, media,
database dumps, and credentials remain outside Git. No public launch or complete
pilot qualification is claimed by this report.
