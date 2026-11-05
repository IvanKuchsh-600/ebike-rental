ALTER TABLE bikes ADD COLUMN status TEXT NOT NULL DEFAULT 'available';

UPDATE bikes SET status = 'rented' WHERE is_rented = TRUE;
UPDATE bikes SET status = 'broken' WHERE is_broken = TRUE;

ALTER TABLE bikes DROP COLUMN is_rented;
ALTER TABLE bikes DROP COLUMN is_broken;

ALTER TABLE bikes ADD CONSTRAINT chk_bike_status
    CHECK (status IN ('available', 'rented', 'broken', 'assembly'));