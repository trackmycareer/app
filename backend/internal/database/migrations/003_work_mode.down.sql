ALTER TABLE jobs ADD COLUMN remote BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE jobs SET remote = TRUE WHERE work_mode = 'remote';
UPDATE jobs SET remote = FALSE WHERE work_mode != 'remote';

ALTER TABLE jobs DROP COLUMN work_mode;
