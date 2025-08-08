-- down.sql
ALTER TABLE documents 
ALTER COLUMN content TYPE text USING content::text;