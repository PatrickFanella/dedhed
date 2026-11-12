# Dedhed roadmap

Status: the owner confirmed preview access and sound quality and approved a two-recording noncommercial public pilot. The approved pilot is live at https://dedhed.subcult.tv; natural daily maintenance and longer stability remain under observation.
Updated: 2026-10-04. Delivery is ordered by evidence gates; no release dates are committed.

Dedhed starts with the Grateful Dead music collection at Internet Archive. V1
combines a browsable concert library, on-demand playback of individual sets and
complete shows, and a shared radio schedule. Playing a chosen set affects only
that listener. Tune in joins the station at its current program position.

[Roadmap tracking issue](https://github.com/subculture-collective/dedhed/issues/4) · [All issues](https://github.com/subculture-collective/dedhed/issues) · [Milestones](https://github.com/subculture-collective/dedhed/milestones)

The [design proposal](docs/plans/0001-grateful-dead-radio.md) records source
observations and playback assumptions. Issues below own executable scope,
acceptance criteria, and dependencies.

The [private pilot evidence report](docs/reports/private-pilot-2026-10-02.md)
separates local and hosted checks, source comparison, browser playback, sustained
delivery, and recovery. [Issue #19](https://github.com/subculture-collective/dedhed/issues/19)
tracks the owner-reported sound-quality concern and an alternate audience source.

## Delivery milestones

| Milestone | Result and exit gate |
| --- | --- |
| [M1 — Foundations and reviewed sources](https://github.com/subculture-collective/dedhed/milestone/1) | Establish the development workflow, Arkhiv reuse boundary, bounded metadata inventory, and a reviewed source sample. Exit: repeatable discovery and explicit source decisions without bulk acquisition. |
| [M2 — On-demand concerts and sets](https://github.com/subculture-collective/dedhed/milestone/2) | Browse reviewed concerts and play a selected set or complete show with correct order, seeking, credits, and visible failure states. Exit: rendered playback evidence and auditioned transitions. |
| [M3 — Shared radio](https://github.com/subculture-collective/dedhed/milestone/3) | Add a persisted shared station schedule and Tune in alongside on-demand playback. Exit: independent listener controls, measured join timing, restart recovery, and no overlapping audio when switching modes. |
| [M4 — Private pilot qualification](https://github.com/subculture-collective/dedhed/milestone/4) | Qualify both listening modes, source failure handling, recovery, and resource use. Exit: a complete-concert private run plus reproducible verification and recovery evidence. Public release remains a separate decision. |

## M1 — Foundations and reviewed sources

| Issue | Prerequisites |
| --- | --- |
| [#5 — Define the Arkhiv reuse boundary and audio domain contracts](https://github.com/subculture-collective/dedhed/issues/5) | None |
| [#6 — Bootstrap the Go/React/PostgreSQL application and isolated development workflow](https://github.com/subculture-collective/dedhed/issues/6) | [#5](https://github.com/subculture-collective/dedhed/issues/5) |
| [#7 — Discover and normalize a bounded Grateful Dead recording inventory](https://github.com/subculture-collective/dedhed/issues/7) | [#5](https://github.com/subculture-collective/dedhed/issues/5) |
| [#8 — Review the pilot recordings and record playback, acquisition, and relay decisions](https://github.com/subculture-collective/dedhed/issues/8) | [#7](https://github.com/subculture-collective/dedhed/issues/7) |

## M2 — On-demand concerts and sets

| Issue | Prerequisites |
| --- | --- |
| [#9 — Persist concerts, recordings, sets, tracks, and editorial review state](https://github.com/subculture-collective/dedhed/issues/9) | [#6](https://github.com/subculture-collective/dedhed/issues/6), [#7](https://github.com/subculture-collective/dedhed/issues/7) |
| [#10 — Prove concert audio delivery and choose the playback transport](https://github.com/subculture-collective/dedhed/issues/10) | [#6](https://github.com/subculture-collective/dedhed/issues/6), [#8](https://github.com/subculture-collective/dedhed/issues/8) |
| [#11 — Implement on-demand playback for complete shows and individual sets](https://github.com/subculture-collective/dedhed/issues/11) | [#9](https://github.com/subculture-collective/dedhed/issues/9), [#10](https://github.com/subculture-collective/dedhed/issues/10) |
| [#12 — Build the concert library and recording/set listening pages](https://github.com/subculture-collective/dedhed/issues/12) | [#9](https://github.com/subculture-collective/dedhed/issues/9), [#11](https://github.com/subculture-collective/dedhed/issues/11) |

## M3 — Shared radio

| Issue | Prerequisites |
| --- | --- |
| [#13 — Implement a persisted shared station schedule and now-playing API](https://github.com/subculture-collective/dedhed/issues/13) | [#9](https://github.com/subculture-collective/dedhed/issues/9), [#10](https://github.com/subculture-collective/dedhed/issues/10) |
| [#14 — Add Tune in and safe switching between radio and on-demand listening](https://github.com/subculture-collective/dedhed/issues/14) | [#13](https://github.com/subculture-collective/dedhed/issues/13), [#12](https://github.com/subculture-collective/dedhed/issues/12) |

## M4 — Private pilot qualification

| Issue | Prerequisites |
| --- | --- |
| [#15 — Implement playback operations, integrity checks, and independent recovery](https://github.com/subculture-collective/dedhed/issues/15) | [#11](https://github.com/subculture-collective/dedhed/issues/11), [#13](https://github.com/subculture-collective/dedhed/issues/13) |
| [#16 — Qualify both listening modes, transitions, and source failure behavior](https://github.com/subculture-collective/dedhed/issues/16) | [#14](https://github.com/subculture-collective/dedhed/issues/14) |
| [#17 — Run and document a complete-concert private library and radio pilot](https://github.com/subculture-collective/dedhed/issues/17) | [#15](https://github.com/subculture-collective/dedhed/issues/15), [#16](https://github.com/subculture-collective/dedhed/issues/16) |

## Show index (added 2026-10-02)

A collection-wide, metadata-only inventory (`source_items`) backs the calendar,
date pages, On this day and full-text search over venues, places, credits and
song titles. It extends issue #7's bounded discovery to the whole collection
without acquiring audio. Date pages mark which source recordings are in the
library and playable; the rest are listed with credits and a source link. The
operator commands are in the [pilot runbook](docs/runbooks/private-pilot.md).

## After private qualification

[#18 — Decide public release scope, distribution permissions, and hosting](https://github.com/subculture-collective/dedhed/issues/18)
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

GitHub issues are authoritative for current task state and dependencies. This
document preserves milestone scope and ordering; update it when scope changes.
The tracking issue checklist should be updated as work is qualified.

## Two-recording public pilot, October 4

The [public pilot runbook](docs/runbooks/public-pilot.md) records the approved
Fillmore East and Barton Hall scope. Public approvals are separate from private
reviews, bind source and media hashes, and filter every audience-facing catalog,
search/calendar, audio and station lookup. Daily renewal verifies the exact
approved audio before replacing the schedule. Deployment, hosted checks,
independent public restore passed. Natural daily runs remain pending. See the [release evidence](docs/reports/public-pilot-2026-10-04.md).

## Audio viewer follow-up

Issue #22 adds volume/mute persistence, accessible narrow controls, original-date
and venue filters before pagination, tape-rack/table browsing, local favorites,
sharing and listening focus. Radio has read-only position, measured replay states
and fresh current-program resume/rejoin. On-demand keeps seeking and reviewed
set/track controls. Viewer actions never write the shared timeline.
