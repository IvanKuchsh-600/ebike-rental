DROP TABLE IF EXISTS payouts;

ALTER TABLE rentals
DROP COLUMN IF EXISTS is_deposit_returned,
    DROP COLUMN IF EXISTS deposit_returned,
    DROP COLUMN IF EXISTS deposit,
    DROP COLUMN IF EXISTS prepayment;