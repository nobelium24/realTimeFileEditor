-- up.sql
ALTER TABLE documents 
ALTER COLUMN content TYPE jsonb USING content::jsonb;