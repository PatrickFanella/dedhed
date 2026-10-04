# Shared audio controls

The Tape Shelf interface pins `@subcult/archive-viewer/audio` 0.2.1 for transport
controls, track timing and reviewed set bounds. The immutable package receipt is
`docs/releases/archive-viewer-0.2.1.json`.

Dedhed owns its catalog, calendar, cassette interface, recording URLs, radio
schedule, credits and access policy. Its source and review records remain in the
existing database. This adoption changes frontend components; it does not change
the backend, approve relay or activate a new schedule.

Run the isolated `make verify` gate before proposing an upgrade. Real music
listening, continuous playback, private recovery and hosted CI are separate
qualification checks. Public deployment requires its existing owner decision.
