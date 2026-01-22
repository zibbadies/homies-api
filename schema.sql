CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SEQUENCE house_seq;
CREATE TABLE houses (
  id INTEGER PRIMARY KEY NOT NULL DEFAULT nextval('house_seq'),
  name VARCHAR(255) NOT NULL,
  owner UUID NOT NULL,
  invite VARCHAR(6) NOT NULL
);

CREATE TABLE users (
  id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
  name VARCHAR(32) UNIQUE NOT NULL,
  house INTEGER DEFAULT NULL,
  pwd_hash BYTEA NOT NULL,
  pwd_salt BYTEA NOT NULL,
  bg_color CHAR(6) NOT NULL,
  face_color CHAR(6) NOT NULL,
  face_x FLOAT NOT NULL,
  face_y FLOAT NOT NULL,
  left_eye_x FLOAT NOT NULL,
  left_eye_y FLOAT NOT NULL,
  right_eye_x FLOAT NOT NULL,
  right_eye_y FLOAT NOT NULL,
  bezier CHAR(11) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
);

CREATE TABLE lists (
  id SERIAL PRIMARY KEY NOT NULL,
  house_id INTEGER NOT NULL,
  name VARCHAR(256) NOT NULL,
  items INT NOT NULL DEFAULT 0
  in_overview BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE todos (
  id SERIAL PRIMARY KEY NOT NULL,
  list_id SERIAL NOT NULL,
  completed BOOLEAN NOT NULL DEFAULT false,
  text VARCHAR(512) NOT NULL,
  author UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
);

CREATE UNIQUE INDEX invite_index ON houses (invite);

CREATE UNIQUE INDEX name_index ON users (name);

CREATE INDEX house_index ON users (house);

CREATE INDEX house_id_index ON lists (house_id);

CREATE INDEX list_id_index ON todos (list_id);

ALTER TABLE users ADD CONSTRAINT link_user_house FOREIGN KEY (house) REFERENCES houses(id) ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE lists ADD CONSTRAINT link_list_house FOREIGN KEY (house_id) REFERENCES houses(id) ON DELETE CASCADE;

ALTER TABLE todos ADD CONSTRAINT link_todo_lists FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE;
