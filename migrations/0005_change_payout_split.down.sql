ALTER TABLE payouts
DROP COLUMN IF EXISTS amortization_share,
    DROP COLUMN IF EXISTS owner_share;