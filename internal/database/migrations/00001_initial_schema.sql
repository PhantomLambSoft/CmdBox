CREATE TABLE profiles
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT     NOT NULL UNIQUE,
    description TEXT     NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    last_used   DATETIME
);

CREATE TABLE profile_states
(
    id                         INTEGER PRIMARY KEY AUTOINCREMENT,
    active_command_profile_id  INTEGER NOT NULL REFERENCES profiles (id) ON DELETE RESTRICT,
    active_variable_profile_id INTEGER NOT NULL REFERENCES profiles (id) ON DELETE RESTRICT,
    active_settings_profile_id INTEGER NOT NULL REFERENCES profiles (id) ON DELETE RESTRICT
);

CREATE TABLE tags
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT     NOT NULL UNIQUE,
    description TEXT     NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE TABLE commands
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    alias       TEXT     NOT NULL,
    template    TEXT     NOT NULL,
    description TEXT     NOT NULL DEFAULT '',
    cwd         TEXT,
    shell       TEXT,
    env         TEXT,
    timeout     INTEGER,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    used        INTEGER  NOT NULL DEFAULT 0,
    last_used   DATETIME,
    profile_id  INTEGER  NOT NULL REFERENCES profiles (id) ON DELETE CASCADE,
    UNIQUE (alias, profile_id)
);

CREATE TABLE variables
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT     NOT NULL,
    value      TEXT     NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    profile_id INTEGER  NOT NULL REFERENCES profiles (id) ON DELETE CASCADE,
    UNIQUE (name, profile_id)
);

CREATE TABLE command_tags
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    command_id INTEGER  NOT NULL REFERENCES commands (id) ON DELETE CASCADE,
    tag_id     INTEGER  NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL,
    UNIQUE (command_id, tag_id)
);

CREATE TABLE variable_tags
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    variable_id INTEGER  NOT NULL REFERENCES variables (id) ON DELETE CASCADE,
    tag_id      INTEGER  NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    created_at  DATETIME NOT NULL,
    UNIQUE (variable_id, tag_id)
);

CREATE TABLE command_histories
(
    id             TEXT PRIMARY KEY,
    alias          TEXT     NOT NULL,
    template       TEXT     NOT NULL,
    resolved       TEXT     NOT NULL,
    variables_used TEXT,
    exit_code      INTEGER,
    ran_at         DATETIME NOT NULL,
    profile_id     INTEGER  NOT NULL REFERENCES profiles (id) ON DELETE CASCADE
);
CREATE INDEX idx_command_histories_alias ON command_histories (alias);
CREATE INDEX idx_command_histories_ran_at ON command_histories (ran_at);
CREATE INDEX idx_command_histories_profile_id ON command_histories (profile_id);

INSERT INTO profiles (name, description, created_at, updated_at, last_used)
VALUES ('default', 'Automatically created default profile.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL);

INSERT INTO profile_states (active_command_profile_id, active_variable_profile_id, active_settings_profile_id)
VALUES (1, 1, 1);

DROP TABLE command_histories;
DROP TABLE variable_tags;
DROP TABLE command_tags;
DROP TABLE variables;
DROP TABLE commands;
DROP TABLE tags;
DROP TABLE profile_states;
DROP TABLE profiles;