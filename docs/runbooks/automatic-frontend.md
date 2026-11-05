# Automatic public frontend delivery

Native Gitea CI retains `make verify` with this run's disposable PostgreSQL 17
service. The frontend job depends directly on that verification job. It builds
and validates the deployable image on PRs without activating production. Main
pushes and manual main dispatches publish the exact frontend and existing API
role image, then activate only the current public frontend on Almaz.

The image explicitly uses `VITE_DEDHED_AUDIENCE=public`, preserving the public
listening/trading notice. The default Docker build remains private for local
use. The maintained nginx 1.31.7 pin retains the live fleet runtime version;
a full rebuild must not revert the older 1.29 source default.

The scoped host baseline selects Compose service `web`, container
`dedhed-public-20261004-web-1`, retaining its existing catalog/media networks,
loopback port, resources and runtime overrides. The pinned Arkhiv core, catalog,
media, access decisions and radio maintenance timers keep the
[shared-core release procedure](shared-core-release.md). This workflow performs
no core upgrade, database migration, acquisition or additional public relay.

An app-specific restricted fleet-deploy key accepts only this repository/service
and current main. The receiver refuses runtime drift, verifies image identity,
health and exact public HTML, and restores the previous image on failed
readiness. Retries reuse the accepted image and skip rebuilding/publication.
Main-only credential guards are step-level so PR image builds actually execute.

Gitea stores `HOMELAB_DEPLOY_KEY`, pinned `HOMELAB_KNOWN_HOSTS`, and the existing
registry secret. The homelab operator owns the deployment key; rotate it by
2027-02-03. Register and test the host profile before merge. Private receipts
retain image IDs, diagnostics and rollback commands. After rollout, compare
public HTML/branding/build files and check the existing catalog/radio endpoints.
Deployment checks are separate from real audio playback qualification.
