-- +goose Up
SET FOREIGN_KEY_CHECKS = 0;
TRUNCATE TABLE notes;
TRUNCATE TABLE note_revisions;
TRUNCATE TABLE user_settings;
TRUNCATE TABLE note_views;
SET FOREIGN_KEY_CHECKS = 1;

ALTER TABLE notes MODIFY created_at INT;
ALTER TABLE notes MODIFY deleted_at INT;
ALTER TABLE notes MODIFY updated_at INT;
ALTER TABLE note_revisions MODIFY updated_at INT;
ALTER TABLE note_views MODIFY viewed_at INT;
