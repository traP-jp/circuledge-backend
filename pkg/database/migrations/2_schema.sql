-- +goose Up

-- notesテーブル（ノートのメタ情報）
CREATE TABLE IF NOT EXISTS notes (
    id VARCHAR(36) NOT NULL, -- UUIDv7
    latest_revision VARCHAR(36) NOT NULL, -- UUIDv7
    created_at DATETIME NOT NULL,
    deleted_at DATETIME DEFAULT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id)
);

-- note_revisionsテーブル（ノートのリビジョン履歴）
CREATE TABLE IF NOT EXISTS note_revisions (
    note_id VARCHAR(36) NOT NULL, -- UUIDv7
    revision_id VARCHAR(36) NOT NULL, -- UUIDv7
    channel VARCHAR(36) NOT NULL, -- UUID
    permission ENUM('public', 'limited', 'editable', 'freely', 'locked', 'private') NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    body TEXT NOT NULL,
    updated_at INT NOT NULL,
    PRIMARY KEY (revision_id),
    INDEX idx_note_id (note_id),
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE -- notesが削除された場合、リビジョンも削除されるようにする
);

-- user_settingsテーブル（ユーザー設定）
CREATE TABLE IF NOT EXISTS user_settings (
    user_name VARCHAR(255) NOT NULL,
    default_channel VARCHAR(36) NOT NULL, -- UUID
    PRIMARY KEY (user_name) -- user_nameを主キーに設定
);

-- note_viewsテーブル（閲覧履歴）
CREATE TABLE IF NOT EXISTS note_views (
    user_name VARCHAR(255) NOT NULL,
    note_id VARCHAR(36) NOT NULL,
    viewed_at DATETIME NOT NULL,
    PRIMARY KEY (user_name, note_id, viewed_at)
);

-- +goose Down
DROP TABLE IF EXISTS note_views;
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS note_revisions;
DROP TABLE IF EXISTS notes;