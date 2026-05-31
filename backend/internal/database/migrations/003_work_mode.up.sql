ALTER TABLE jobs ADD COLUMN work_mode VARCHAR(20) NOT NULL DEFAULT 'onsite';

UPDATE jobs SET work_mode = 'remote' WHERE remote = TRUE;
UPDATE jobs SET work_mode = 'onsite' WHERE remote = FALSE;

ALTER TABLE jobs DROP COLUMN remote;
