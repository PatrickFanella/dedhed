# Development and verification

Use Go 1.26+, Node 24, pnpm 10.33.0, and PostgreSQL 17. The T3 `dedhed` profile
provides isolated dependency caches, source snapshots, preview media, and
separate test database state. Read the shared T3 agent guide before using it.

```sh
t3-dev --profile dedhed start
t3-dev --profile dedhed sync
t3-dev --profile dedhed verify
t3-dev --profile dedhed stop
```

Start seeds synthetic test tones and applies checksum-locked migrations. Seeding
does not replace an existing station schedule. The UI labels these fixtures;
they are not actual concerts. Application/test networks cannot acquire music.

`make verify` checks pinned source provenance, local document links, Python
acquisition policy tests, Go formatting/vet/race tests/build, TypeScript, ESLint,
web tests, and the production web build. Database integration tests run only
with an explicitly disposable `TEST_DATABASE_URL`; the launcher supplies it.
The Gitea workflow uses its own PostgreSQL 17 service. Hosted CI and rendered
playback qualification remain separate proof.

For manual setup, use the nonsecret `.env.example` values with a disposable DB,
install web dependencies, and download Go modules. From `backend/`:

```sh
go run ./cmd/dedhed migrate
go run ./cmd/dedhed seed
go run ./cmd/dedhed serve
```

Then run `pnpm --dir web dev`. Vite forwards `/api` to loopback port 8080;
`DEDHED_API` overrides that origin for a separately isolated private pilot.
Server health is `/api/health`, database readiness is `/api/ready`, and listening
behavior requires `/api/recordings`, `/api/audio/{id}`, and `/api/station` checks.
Public HTTP routes are read-only. Operator mutations use the CLI.

See the [private operations runbook](docs/runbooks/private-pilot.md) for source
review, acquisition, station scheduling, and recovery. Keep generated audio and
state outside Git. Update [source provenance](backend/third_party/arkhiv/provenance.json)
only through a reviewed source-package update.
