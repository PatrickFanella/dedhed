# Automatic public frontend delivery

Pull requests run `.github/workflows/mirror-ci.yml` on GitHub: `make verify`
with the run's disposable PostgreSQL 17 service, plus a job that builds and
validates the deployable image without activating production. Main pushes and
manual main dispatches run `.github/workflows/deploy.yml`, which joins the
tailnet, publishes the exact frontend and existing API role image, then
activates only the current public frontend on Almaz. The receiver deploys only
the revision that Gitea's copy of `main` has reached, so the job waits for it.

The image explicitly uses `VITE_DEDHED_AUDIENCE=public`, preserving the public
listening/trading notice. The default Docker build remains private for local
use. The nginx base image in `web/Dockerfile` is pinned by tag and digest;
a full rebuild must not revert to the older 1.29 source default.

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

GitHub stores `HOMELAB_DEPLOY_KEY`, pinned `HOMELAB_KNOWN_HOSTS`, and the
Tailscale OAuth client in `TS_OAUTH_CLIENT_ID` and `TS_OAUTH_SECRET`; images go
to GitHub Packages with the workflow token. Almaz accepts the key only from
tailnet addresses. The homelab operator owns the deployment key; rotate it by
2027-02-03. Register and test the host profile before merge. Private receipts
retain image IDs, diagnostics and rollback commands. After rollout, compare
public HTML/branding/build files and check the existing catalog/radio endpoints.
Deployment checks are separate from real audio playback qualification.
