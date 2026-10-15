ALTER TABLE bikes
DROP COLUMN IF EXISTS photo_path,
    DROP COLUMN IF EXISTS photo_updated_at;