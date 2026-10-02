# Dedhed roadmap

Status: private implementation under verification. Pilot qualification is in progress; public deployment remains deferred.
Updated: 2026-10-02. Delivery is ordered by evidence gates; no release dates are committed.

Dedhed starts with the Grateful Dead music collection at Internet Archive. V1
combines a browsable concert library, on-demand playback of individual sets and
complete shows, and a shared radio schedule. Playing a chosen set affects only
that listener. Tune in joins the station at its current program position.

[Roadmap tracking issue](https://git.subcult.tv/subculture-collective/dedhed/issues/1) · [All issues](https://git.subcult.tv/subculture-collective/dedhed/issues) · [Milestones](https://git.subcult.tv/subculture-collective/dedhed/milestones)

The [design proposal](docs/plans/0001-grateful-dead-radio.md) records source
observations and playback assumptions. Issues below own executable scope,
acceptance criteria, and native Gitea dependencies.

## Delivery milestones

| Milestone | Result and exit gate |
| --- | --- |
| [M1 — Foundations and reviewed sources](https://git.subcult.tv/subculture-collective/dedhed/milestone/94) | Establish the development workflow, Arkhiv reuse boundary, bounded metadata inventory, and a reviewed source sample. Exit: repeatable discovery and explicit source decisions without bulk acquisition. |
| [M2 — On-demand concerts and sets](https://git.subcult.tv/subculture-collective/dedhed/milestone/95) | Browse reviewed concerts and play a selected set or complete show with correct order, seeking, credits, and visible failure states. Exit: rendered playback evidence and auditioned transitions. |
| [M3 — Shared radio](https://git.subcult.tv/subculture-collective/dedhed/milestone/96) | Add a persisted shared station schedule and Tune in alongside on-demand playback. Exit: independent listener controls, measured join timing, restart recovery, and no overlapping audio when switching modes. |
| [M4 — Private pilot qualification](https://git.subcult.tv/subculture-collective/dedhed/milestone/97) | Qualify both listening modes, source failure handling, recovery, and resource use. Exit: a complete-concert private run plus reproducible verification and recovery evidence. Public release remains a separate decision. |

## M1 — Foundations and reviewed sources

| Issue | Prerequisites |
| --- | --- |
| [#2 — Define the Arkhiv reuse boundary and audio domain contracts](https://git.subcult.tv/subculture-collective/dedhed/issues/2) | None |
| [#3 — Bootstrap the Go/React/PostgreSQL application and isolated development workflow](https://git.subcult.tv/subculture-collective/dedhed/issues/3) | [#2](https://git.subcult.tv/subculture-collective/dedhed/issues/2) |
| [#4 — Discover and normalize a bounded Grateful Dead recording inventory](https://git.subcult.tv/subculture-collective/dedhed/issues/4) | [#2](https://git.subcult.tv/subculture-collective/dedhed/issues/2) |
| [#5 — Review the pilot recordings and record playback, acquisition, and relay decisions](https://git.subcult.tv/subculture-collective/dedhed/issues/5) | [#4](https://git.subcult.tv/subculture-collective/dedhed/issues/4) |

## M2 — On-demand concerts and sets

| Issue | Prerequisites |
| --- | --- |
| [#6 — Persist concerts, recordings, sets, tracks, and editorial review state](https://git.subcult.tv/subculture-collective/dedhed/issues/6) | [#3](https://git.subcult.tv/subculture-collective/dedhed/issues/3), [#4](https://git.subcult.tv/subculture-collective/dedhed/issues/4) |
| [#7 — Prove concert audio delivery and choose the playback transport](https://git.subcult.tv/subculture-collective/dedhed/issues/7) | [#3](https://git.subcult.tv/subculture-collective/dedhed/issues/3), [#5](https://git.subcult.tv/subculture-collective/dedhed/issues/5) |
| [#8 — Implement on-demand playback for complete shows and individual sets](https://git.subcult.tv/subculture-collective/dedhed/issues/8) | [#6](https://git.subcult.tv/subculture-collective/dedhed/issues/6), [#7](https://git.subcult.tv/subculture-collective/dedhed/issues/7) |
| [#9 — Build the concert library and recording/set listening pages](https://git.subcult.tv/subculture-collective/dedhed/issues/9) | [#6](https://git.subcult.tv/subculture-collective/dedhed/issues/6), [#8](https://git.subcult.tv/subculture-collective/dedhed/issues/8) |

## M3 — Shared radio

| Issue | Prerequisites |
| --- | --- |
| [#10 — Implement a persisted shared station schedule and now-playing API](https://git.subcult.tv/subculture-collective/dedhed/issues/10) | [#6](https://git.subcult.tv/subculture-collective/dedhed/issues/6), [#7](https://git.subcult.tv/subculture-collective/dedhed/issues/7) |
| [#11 — Add Tune in and safe switching between radio and on-demand listening](https://git.subcult.tv/subculture-collective/dedhed/issues/11) | [#10](https://git.subcult.tv/subculture-collective/dedhed/issues/10), [#9](https://git.subcult.tv/subculture-collective/dedhed/issues/9) |

## M4 — Private pilot qualification

| Issue | Prerequisites |
| --- | --- |
| [#12 — Implement playback operations, integrity checks, and independent recovery](https://git.subcult.tv/subculture-collective/dedhed/issues/12) | [#8](https://git.subcult.tv/subculture-collective/dedhed/issues/8), [#10](https://git.subcult.tv/subculture-collective/dedhed/issues/10) |
| [#13 — Qualify both listening modes, transitions, and source failure behavior](https://git.subcult.tv/subculture-collective/dedhed/issues/13) | [#11](https://git.subcult.tv/subculture-collective/dedhed/issues/11) |
| [#14 — Run and document a complete-concert private library and radio pilot](https://git.subcult.tv/subculture-collective/dedhed/issues/14) | [#12](https://git.subcult.tv/subculture-collective/dedhed/issues/12), [#13](https://git.subcult.tv/subculture-collective/dedhed/issues/13) |

## After private qualification

[#15 — Decide public release scope, distribution permissions, and hosting](https://git.subcult.tv/subculture-collective/dedhed/issues/15)
records source use, domain/host, audience, monetization, capacity, and rollout
decisions. It does not authorize or perform a public deployment.

## Product and source boundaries

- Both on-demand sets/full shows and shared radio are required for private v1.
- Sets are reviewed track groups; disc splits are not automatically set boundaries.
- Preserve recording credits, source links, track order, ambience, and segues.
- Arkhiv reuse requires an explicit versioned boundary; audio does not become a PDF document variant.
- Source availability, acquisition, on-demand playback, and station relay are independent decisions.
- Begin with a small reviewed audience-recording sample. Soundboards stay outside local acquisition and relay in this milestone.
- Audio delivery remains an evidence-based choice: on-demand requires seeking; radio requires measured continuity and shared timing.
- Accounts, custom personal queues, requests, favorites, other bands, and DJ commentary are deferred.

## Evidence and maintenance

Issues are open planning records, not claims of implementation. Local tests,
hosted CI, rendered browser checks, auditioned source playback, private runtime
qualification, restore proof, and public deployment are separate evidence.
Attach the relevant receipts before closing an issue. A failed acceptance check
keeps that issue open even if its code has merged.

Gitea issues are authoritative for current task state and dependencies. This
document preserves milestone scope and ordering; update it when scope changes.
The tracking issue checklist should be updated as work is qualified.
