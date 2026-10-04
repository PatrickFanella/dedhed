CREATE TABLE public_approvals (
  recording_id text PRIMARY KEY REFERENCES recordings(id),
  approved boolean NOT NULL,
  source_hash text NOT NULL,
  media_sha256 text NOT NULL,
  policy_basis text NOT NULL,
  reviewer text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE public_approval_history (
  id bigserial PRIMARY KEY,
  recording_id text NOT NULL REFERENCES recordings(id),
  approval jsonb NOT NULL,
  recorded_at timestamptz NOT NULL DEFAULT now()
);
