CREATE SCHEMA IF NOT EXISTS dating_data;

-- DROP SCHEMA dating_data cascade;

CREATE TABLE IF NOT EXISTS dating_data.user (
    user_id SERIAL,
    login VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    phone_number VARCHAR(255),
    email VARCHAR(255),
    last_online TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS dating_data.auth (
    token_id SERIAL,
    user_id INTEGER NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    end_time TIMESTAMP NOT NULL,

    PRIMARY KEY (token_id),
    FOREIGN KEY (user_id) REFERENCES dating_data.user
);

CREATE TABLE IF NOT EXISTS dating_data.starred_users (
    user_id INTEGER NOT NULL,
    starred_user_id INTEGER NOT NULL,
    viewed BOOL NOT NULL DEFAULT FALSE,
    is_liked BOOL NOT NULL DEFAULT FALSE,
    time TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (user_id, starred_user_id),
    FOREIGN KEY (user_id) REFERENCES dating_data.user,
    FOREIGN KEY (starred_user_id) REFERENCES dating_data.user
);

CREATE TABLE IF NOT EXISTS dating_data.indexed_users (
    user_id INTEGER NOT NULL,
    indexed_user_id INTEGER NOT NULL,

    PRIMARY KEY (user_id, indexed_user_id),
    FOREIGN KEY (user_id) REFERENCES dating_data.user,
    FOREIGN KEY (indexed_user_id) REFERENCES dating_data.user
);

CREATE TABLE IF NOT EXISTS dating_data.profile (
    profile_id SERIAL,
    user_id INTEGER NOT NULL UNIQUE,
    profile_text TEXT NOT NULL,
    sex BOOL NOT NULL,
    birthday TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    url TEXT,

    PRIMARY KEY (profile_id),
    FOREIGN KEY (user_id) REFERENCES  dating_data.user
);

CREATE TABLE IF NOT EXISTS dating_data.profile_history (
    profile_id INTEGER NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    profile_text TEXT NOT NULL,
    sex BOOL NOT NULL,
    birthday TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    url TEXT,
    end_time TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (profile_id),
    FOREIGN KEY (user_id) REFERENCES  dating_data.user
);

CREATE TABLE IF NOT EXISTS dating_data.report (
    user_id INTEGER NOT NULL,
    profile_id INTEGER NOT NULL,
    message VARCHAR NOT NULL,

    PRIMARY KEY (user_id, profile_id),
    FOREIGN KEY (user_id) REFERENCES dating_data.user (user_id),
    FOREIGN KEY (profile_id) REFERENCES dating_data.profile (profile_id)
);

CREATE TABLE IF NOT EXISTS dating_data.report_history (
    user_id INTEGER NOT NULL,
    profile_id INTEGER NOT NULL,
    message TEXT NOT NULL,

    PRIMARY KEY (user_id, profile_id),
    FOREIGN KEY (user_id) REFERENCES dating_data.user (user_id),
    FOREIGN KEY (profile_id) REFERENCES dating_data.profile_history (profile_id)
);

CREATE TABLE IF NOT EXISTS dating_data.photo (
    photo_id SERIAL,
    image_url TEXT NOT NULL,

    PRIMARY KEY (photo_id)
);

CREATE TABLE IF NOT EXISTS dating_data.profile_photo (
    profile_id INTEGER NOT NULL,
    photo_id INTEGER NOT NULL,

    PRIMARY KEY (profile_id, photo_id),
    FOREIGN KEY (profile_id) REFERENCES dating_data.profile (profile_id),
    FOREIGN KEY (photo_id) REFERENCES dating_data.photo (photo_id)
);

CREATE TABLE IF NOT EXISTS dating_data.profile_photo_history (
    profile_id INTEGER NOT NULL,
    photo_id INTEGER NOT NULL,

    PRIMARY KEY (profile_id, photo_id),
    FOREIGN KEY (profile_id) REFERENCES dating_data.profile_history (profile_id),
    FOREIGN KEY (photo_id) REFERENCES dating_data.photo (photo_id)
);

CREATE TABLE IF NOT EXISTS dating_data.photo_report (
    user_id INTEGER NOT NULL,
    photo_id INTEGER NOT NULL,
    message TEXT NOT NULL,

    PRIMARY KEY (user_id, photo_id),
    FOREIGN KEY (user_id) REFERENCES dating_data.user (user_id),
    FOREIGN KEY (user_id) REFERENCES dating_data.photo (photo_id)
);
