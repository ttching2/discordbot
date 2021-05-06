PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users(
    users_id INTEGER PRIMARY KEY,
    discord_users_id BIG INTEGER NOT NULL UNIQUE,
    user_name TEXT,
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

CREATE TABLE IF NOT EXISTS tournament(
    tournament_id INTEGER PRIMARY KEY,
    author INTEGER,
    challonge_id TEXT,
    discord_server_id BIG INTEGER UNIQUE,
    current_match INTEGER,
    FOREIGN KEY(author) REFERENCES users(users_id)
);

CREATE TABLE IF NOT EXISTS tournament_participant(
    tournament_participant_id INTEGER PRIMARY KEY,
    name TEXT,
    challonge_id INTEGER UNIQUE
);

CREATE TABLE IF NOT EXISTS tournament_organizer_xref(
    tournament_id INTEGER,
    users_id INTEGER,
    FOREIGN KEY(tournament_id) REFERENCES tournament(tournament_id) ON DELETE CASCADE,
    FOREIGN KEY(users_id) REFERENCES users(users_id),
    PRIMARY KEY(tournament_id, users_id)
);

CREATE TABLE IF NOT EXISTS tournament_participant_xref(
    tournament_id INTEGER,
    tournament_participant_id INTEGER,
    FOREIGN KEY(tournament_id) REFERENCES tournament(tournament_id) ON DELETE CASCADE,
    FOREIGN KEY(tournament_participant_id) REFERENCES tournament_participant(tournament_participant_id)
    PRIMARY KEY(tournament_id, tournament_participant_id)
);