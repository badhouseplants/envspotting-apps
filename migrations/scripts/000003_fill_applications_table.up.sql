DO $$ 
  BEGIN
    BEGIN
      ALTER TABLE applications ADD COLUMN id TEXT PRIMARY KEY;
      ALTER TABLE applications ADD COLUMN name TEXT;
      ALTER TABLE applications ADD COLUMN description TEXT;
      ALTER TABLE applications ADD COLUMN contours TEXT[];
    EXCEPTION
      WHEN duplicate_column THEN RAISE NOTICE 'column already exists in applications.';
    END;
  END;
$$;
