-- Add includes_gst column to sessions table
ALTER TABLE sessions ADD COLUMN includes_gst BOOLEAN DEFAULT 0 NOT NULL;