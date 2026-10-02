# Dedhed

A Grateful Dead concert library with on-demand sets and a shared radio station,
using source tools from [Arkhiv](https://git.subcult.tv/subculture-collective/arkhiv).

Browse reviewed recordings, play an individual set or full show, or **Tune in**
to the station's current program. On-demand controls affect only that listener.
The station keeps its own timeline. Recording credits and source links stay
with each tape.

The first private implementation includes a PostgreSQL catalog, bounded metadata
inspection, reviewed audience acquisition, continuous audio derivatives, set and
track seeking, a persisted station schedule, and backup/restore tools. The T3
preview uses clearly labeled synthetic test tones. Real source playback and pilot
qualification are separate; public deployment and relay are not enabled.

See [DEVELOPMENT.md](DEVELOPMENT.md) for setup and verification, the
[roadmap](ROADMAP.md) for milestones and issues, and the
[design proposal](docs/plans/0001-grateful-dead-radio.md) for source evidence and
product decisions. Private operations are in the
[pilot runbook](docs/runbooks/private-pilot.md).

Track current work in [Gitea issues](https://git.subcult.tv/subculture-collective/dedhed/issues)
and [milestones](https://git.subcult.tv/subculture-collective/dedhed/milestones).
Source code is GPL-3.0-only; this license does not apply to acquired audio.
