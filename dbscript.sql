PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users(
    users_id INTEGER PRIMARY KEY,
    discord_users_id BIG INTEGER NOT NULL UNIQUE,
    is_admin BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS in_progress_role_command(
    in_progress_role_command_pk INTEGER PRIMARY KEY,
    guild BIG INT,
    origin_channel BIG INT UNIQUE,
    target_channel BIG INT,
    user BIG INT UNIQUE,
    role BIG INT,
    emoji BIG INT,
    stage INTEGER);

CREATE TABLE IF NOT EXISTS role_message_command(
    role_message_command_pk INTEGER PRIMARY KEY,
    author INTEGER,
    guild BIG INTEGER NOT NULL,
    msg BIG INTEGER NOT NULL UNIQUE,
    role BIG INTEGER,
    emoji BIG INTEGER,
    FOREIGN KEY(author) REFERENCES users(users_id));

CREATE TABLE IF NOT EXISTS twitter_follow_command(
    twitter_follow_command_id INTEGER PRIMARY KEY,
    author INTEGER,
    screen_name TEXT,
    channel BIG INTEGER,
    guild BIG INTEGER, 
    screen_name_id TEXT,
    FOREIGN KEY(author) REFERENCES users(users_id));

CREATE TABLE IF NOT EXISTS strawpoll_deadline(
    strawpoll_deadline_id INTEGER PRIMARY KEY,
    author INTEGER,
    strawpoll_id TEXT,
    guild BIG INTEGER,
    channel BIG INTEGER,
    role BIG INTEGER,
    FOREIGN KEY(author) REFERENCES users(users_id));