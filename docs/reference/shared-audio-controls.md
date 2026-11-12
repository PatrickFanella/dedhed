# Shared audio controls

The Tape Shelf interface uses `@subcult/archive-viewer/audio` for transport
controls, track timing and reviewed set bounds. The pinned version and its
package receipt are in `deploy/viewer-release.json`, checked by
`scripts/check_viewer_pin.py`.

Dedhed owns its catalog, calendar, cassette interface, recording URLs, radio
schedule, credits and access policy. Its source and review records remain in the
existing database. This adoption changes frontend components; it does not change
the backend, approve relay or activate a new schedule.

Run the isolated `make verify` gate before proposing an upgrade. Real music
listening, continuous playback, private recovery and hosted CI are separate
qualification checks. Public deployment requires its existing owner decision.
