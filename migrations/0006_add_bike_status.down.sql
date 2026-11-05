ALTER TABLE bikes ADD COLUMN is_rented BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE bikes ADD COLUMN is_broken BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE bikes SET is_rented = TRUE WHERE status = 'rented';
UPDATE bikes SET is_broken = TRUE WHERE status = 'broken';

ALTER TABLE bikes DROP CONSTRAINT chk_bike_status;
ALTER TABLE bikes DROP COLUMN status;