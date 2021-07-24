DO $$ 
  BEGIN
    BEGIN
      ALTER TABLE contours ADD COLUMN id TEXT PRIMARY KEY;
      ALTER TABLE contours ADD COLUMN application_id TEXT REFERENCES applications(id) ON DELETE CASCADE;
      ALTER TABLE contours ADD COLUMN password TEXT;
      ALTER TABLE contours ADD COLUMN name TEXT;
      ALTER TABLE contours ADD COLUMN description TEXT;
      ALTER TABLE contours ADD COLUMN services JSONB;
    EXCEPTION
      WHEN duplicate_column THEN RAISE NOTICE 'column already exists in contours.';
    END;
  END;
$$;

