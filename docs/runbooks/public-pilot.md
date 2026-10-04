# Public audience-recording pilot

The owner approved a free, noncommercial Dedhed pilot on October 4, 2026,
limited to these two audience recordings:

- `gd1969-06-21.aud-early.cotsman.8995.sbeok.shnf`: Jeff Cook's Fillmore East early show, transferred by Jeff Cotsman.
- `gd1977-05-08.ecm33p.moore.berger.miller.117026.flac16`: Jerry Moore's Barton Hall recording, with Rob Berger transfers and Charlie Miller mastering.

The owner confirmed that the private preview works and sounds good. That is an
owner listening result for the preview, not a claim that every track was heard.
Source cuts, patches, uncertain boundaries and taper/transfer credits remain
with each recording. No additional recordings or acquisition are authorized.

The public scope covers on-demand shows and sets and shared scheduled playback.
It excludes ads, paid access, listener-data exploitation and commercial promotion.
The site displays the trading policy and a correction/removal contact. Permission
withdrawal must stop access. This decision applies to this pilot; it does not
change permissions for the other collections or the private original database.

## Policy evidence

Internet Archive's [Grateful Dead collection](https://archive.org/details/GratefulDead)
publishes the band's noncommercial digital-exchange statement and reserves its
withdrawal. Its [collection guide](https://archivesupport.zendesk.com/hc/en-us/articles/360004715891-The-Grateful-Dead-Collection)
distinguishes downloadable audience recordings from stream-only soundboards.
Its [rights guidance](https://archivesupport.zendesk.com/hc/en-us/articles/360014759692-Rights)
does not guarantee an item's copyright status. The release decision uses the
collection statement, these exact audience lineages, current source availability,
and the owner's approved noncommercial scope. It is not a blanket license for
the collection or soundboards.

## Prepare separate state

Restore the independently recovered private checkpoint into a new database.
Keep the original database, audio and schedules. Copy only the two verified
audio files; do not acquire more content. Apply the checksum-locked migration
with the operator role. `public_approvals` and `public_approval_history` retain
new public decisions independently of the original private reviews.

For each recording, prepare a private JSON approval with `approved`,
`source_hash`, `media_sha256`, `policy_basis`, and `reviewer`. The hashes must
match the stored source snapshot and verified playback file. Record the owner
decision and policy URLs in the basis. Use the operator's existing secret file:

```sh
dedhed public-review -id RECORDING_ID -file APPROVAL_JSON
dedhed public-schedule -ids FILLMORE_ID,BARTON_ID -hours 168
```

The schedule command verifies complete media hashes and rejects missing,
withdrawn, changed or duplicate recordings before replacing the schedule.
Run it daily with the exact two IDs. A future approval does not add a recording
to this schedule automatically.

Serve with `DEDHED_AUDIENCE=public`, a SELECT-only database role, and a read-only
media mount. Build the frontend with `VITE_DEDHED_AUDIENCE=public`. The default
build and serving mode remain private. Unknown serving audiences are rejected.
Public catalogs, calendar/search, audio and station lookups use matching public
approvals. Refreshing a source snapshot or replacing its media invalidates the
old approval; restoring it needs a new operator decision.

The optional Arkhiv viewer package remains pinned independently of Dedhed's
backend. This release uses Dedhed's own audio API and does not claim adoption
of the mixed Arkhiv headless backend.

## Qualification and activation

Run `t3-dev --profile dedhed verify` from the exact source worktree. It includes
the disposable PostgreSQL integration checks, source/hash invalidation,
approval history, private catalog exclusion, synthetic-source rejection and
public audio withdrawal. Hosted CI, private/public browser listening, range
delivery, schedule timing and independent recovery are separate checks.

Before ingress activation, verify both complete media hashes, public scope,
the trading notice, original credits, denial of a private recording, actual
playback and advancing server time. Save the image IDs, source revision,
configuration hashes and independent checkpoint. Activate only this site's
ingress block. Verify the HTTPS hostname and actual media through it.

## Withdrawal and recovery

Record a public review with `approved=false`, a reviewer and the withdrawal
basis. Audio and catalog access then fail closed; a scheduled withdrawn show
becomes unavailable and schedule renewal rejects it. Retain history and recovery
evidence. Remove the site's ingress block if the whole pilot must be withdrawn.

Back up the database and both verified media files to an independent host with
matching hashes. Restore into a new isolated database, then prove public filtering,
on-demand range delivery and station behavior from that restored state.
Rollback removes the new public ingress or restores an explicitly qualified
public artifact and checkpoint. Never expose the retained private image as a
public rollback merely because it previously passed private playback checks.

Natural daily renewals and backups and longer stability observation remain
separate from manual runs and immediate deployment checks.
