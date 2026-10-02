# Dedhed agent instructions

Keep Go under `backend/`, React/TypeScript under `web/`, SQL under `migrations/`,
and operational scripts under `scripts/`. Use Go 1.26+ and pnpm 10.33.0.
Read the roadmap and architecture decision before changing domain boundaries.
The pinned Arkhiv package has its own provenance; do not edit its source silently.

Both on-demand sets/full shows and shared radio are required. Preserve credits,
track order, uncertain set boundaries, and independent access decisions. Never
claim source availability authorizes public relay. Acquired audio, snapshots,
derived files, databases, and credentials stay outside Git.

Required gate: `make verify`; database integration: `make test-integration` with
an explicitly disposable `TEST_DATABASE_URL`. Browser checks and real playback
qualification are separate. Read the shared T3 agent guide before previews/tests;
use profile `dedhed`. Live acquisition is a bounded operator workflow outside
isolated development networks. Do not modify other active previews or databases.
