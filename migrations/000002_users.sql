-- Create users table based on YouTrack Users API (User entity)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_type') THEN
    CREATE TYPE user_type AS ENUM ('AGENT', 'STANDARD_USER', 'REPORTER');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS users (
  -- The identifier of the user.
  id               VARCHAR(64) PRIMARY KEY,
  -- The login name of the user.
  login            VARCHAR(255) NOT NULL UNIQUE,
  -- The email address of the user.
  email            VARCHAR(255),
  -- The full name of the user (first and last name).
  full_name        VARCHAR(255),
  -- The name of the user.
  name             VARCHAR(255),
  -- The URL of the user's avatar.
  avatar_url       TEXT,
  -- Indicates whether the user is banned.
  banned           BOOLEAN NOT NULL DEFAULT FALSE,
  -- Indicates whether the user is currently online.
  online           BOOLEAN NOT NULL DEFAULT FALSE,
  -- Indicates if this user is a guest.
  guest            BOOLEAN,
  -- The type of the user. Possible values: AGENT, STANDARD_USER, REPORTER.
  type             user_type,
  -- The ID of the Hub that manages this user account.
  hub_id           VARCHAR(64),
  -- The ID of the Ring that manages this user account.
  ring_id          VARCHAR(64),
  -- The timestamp of the user's last login.
  last_login_time  TIMESTAMPTZ,
  -- The timestamp of the user's creation in YouTrack.
  creation_time    TIMESTAMPTZ,
  -- The timestamp of the user's last update in YouTrack.
  update_time      TIMESTAMPTZ,
  -- Indicates whether the user's personal data has been anonymized.
  is_anonymized    BOOLEAN NOT NULL DEFAULT FALSE,
  -- The reason why the user is banned.
  ban_reason       TEXT,
  -- Hashed password of the user.
  hash_password    TEXT,
  -- Application audit timestamp when this record was created.
  created_at       TIMESTAMPTZ,
  -- Application audit timestamp when this record was last updated.
  updated_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_login ON users (login);
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_banned ON users (banned);