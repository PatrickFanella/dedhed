package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Public approval is independent of the retained private playback review.
// It covers the verified local audio, on-demand playback and scheduled relay.
type PublicApproval struct {
	Approved    bool   `json:"approved"`
	SourceHash  string `json:"source_hash"`
	MediaSHA256 string `json:"media_sha256"`
	PolicyBasis string `json:"policy_basis"`
	Reviewer    string `json:"reviewer"`
}

func (d *DB) publicPredicate() string {
	if !d.Public {
		return "true"
	}
	return `id IN (SELECT r.id FROM recordings r JOIN public_approvals a ON a.recording_id=r.id
 WHERE a.approved AND a.source_hash=r.source_hash AND a.media_sha256=r.media_sha256
 AND r.media<>'' AND r.review->>'published'='true' AND r.review->>'playback'='true'
 AND r.review->>'kind'='audience' AND coalesce(r.source->>'synthetic','false')<>'true')`
}

func (d *DB) ApprovePublic(ctx context.Context, id string, a PublicApproval) error {
	if d.Public {
		return fmt.Errorf("public serving cannot change approvals")
	}
	if strings.TrimSpace(a.PolicyBasis) == "" || strings.TrimSpace(a.Reviewer) == "" {
		return fmt.Errorf("public decision needs reviewer and policy basis")
	}
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var sourceHash, mediaHash, media string
	var review, source []byte
	if err = tx.QueryRow(ctx, `SELECT source_hash,media_sha256,media,review,source FROM recordings WHERE id=$1 FOR UPDATE`, id).Scan(&sourceHash, &mediaHash, &media, &review, &source); err != nil {
		return err
	}
	if a.Approved {
		var r struct {
			Published bool
			Playback  bool
			Kind      string
		}
		var s struct{ Synthetic bool }
		if err = json.Unmarshal(review, &r); err != nil {
			return err
		}
		if err = json.Unmarshal(source, &s); err != nil {
			return err
		}
		if sourceHash != a.SourceHash || mediaHash != a.MediaSHA256 || len(sourceHash) != 64 || len(mediaHash) != 64 || media == "" || !r.Published || !r.Playback || r.Kind != "audience" || s.Synthetic {
			return fmt.Errorf("public approval requires exact reviewed audience source and verified audio")
		}
	}
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO public_approval_history(recording_id,approval) VALUES($1,$2)`, id, b); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO public_approvals(recording_id,approved,source_hash,media_sha256,policy_basis,reviewer) VALUES($1,$2,$3,$4,$5,$6)
 ON CONFLICT(recording_id) DO UPDATE SET approved=excluded.approved,source_hash=excluded.source_hash,media_sha256=excluded.media_sha256,policy_basis=excluded.policy_basis,reviewer=excluded.reviewer,updated_at=now()`, id, a.Approved, a.SourceHash, a.MediaSHA256, a.PolicyBasis, a.Reviewer); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
