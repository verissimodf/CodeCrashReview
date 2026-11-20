-- +goose Up
-- Preserve historical decisions before revoking approvals without a proof.
ALTER TABLE repository_reviews ADD COLUMN previous_verdicts jsonb NOT NULL DEFAULT '[]'::jsonb;

UPDATE repository_reviews r
SET previous_verdicts = r.previous_verdicts || jsonb_build_array(to_jsonb(v)),
    status = 'awaiting_human', completed_at = NULL,
    error_message = 'Historical verdict revoked: no independently verified functional proof. Previous decision preserved in audit history.'
FROM review_verdicts v
WHERE v.review_id = r.id AND NOT EXISTS (
    SELECT 1 FROM review_proposals p WHERE p.review_id = r.id
    AND (p.slot = v.winner OR v.winner = 'TIE') AND p.patch_valid
    AND COALESCE((p.verification->'proof'->>'passed')::boolean, FALSE)
);

DELETE FROM review_verdicts v USING repository_reviews r
WHERE v.review_id = r.id AND jsonb_array_length(r.previous_verdicts) > 0
AND r.status = 'awaiting_human';

-- +goose Down
-- Revoked approvals must not be resurrected by rollback. Keep the audit column.
SELECT 1;
