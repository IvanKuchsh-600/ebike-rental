ALTER TABLE rentals
    ADD COLUMN paid_at TIMESTAMPTZ;

-- Опционально: перенести старые данные.
-- Если is_paid = TRUE, но paid_at NULL, ставим created_at (или started_at)
UPDATE rentals
SET paid_at = created_at
WHERE is_paid = TRUE AND paid_at IS NULL;