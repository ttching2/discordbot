CREATE TABLE IF NOT EXISTS in_progress_role_command(
    in_progress_role_command_pk INTEGER PRIMARY KEY,
    guild UNSIGNED BIG INT,
    channel UNSIGNED BIG INT,
    user UNSIGNED BIG INT UNIQUE,
    role UNSIGNED BIG INT,
    emoji USIGNED BIG INT,
    stage INTEGER);

CREATE TABLE IF NOT EXISTS role_message_command(
    role_message_command_pk INTEGER PRIMARY KEY,
    guild BIG INTEGER NOT NULL,
    msg BIG INTEGER NOT NULL UNIQUE,
    role BIG INTEGER,
    emoji BIG INTEGER);

CREATE TABLE IF NOT EXISTS twitter_follow_command(
    twitter_follow_command_id INTEGER PRIMARY KEY,
    screen_name TEXT,
    channel BIG INTEGER,
    guild BIG INTEGER, 
    screen_name_id TEXT);

CREATE TABLE IF NOT EXISTS strawpoll_deadline(
    strawpoll_deadline_id INTEGER PRIMARY KEY,
    strawpoll_id TEXT,
    guild BIG INTEGER,
    channel BIG INTEGER,
    role BIG INTEGER);