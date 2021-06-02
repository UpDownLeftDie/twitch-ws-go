CREATE TABLE oauth_token (
  client_id TEXT NOT NULL,
  access_token TEXT NOT NULL,
  refresh_token TEXT NOT NULL,
  expires_at TEXT NOT NULL,

  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

  PRIMARY KEY (client_id)
);

CREATE TABLE twitch_viewer(
  id TEXT NOT NULL,
  username TEXT NOT NULL,

  PRIMARY KEY (id)
)

CREATE TABLE events.twitch (
  id TEXT NOT NULL,
  type TEXT NOT NULL,
  amount TEXT,

  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

  PRIMARY KEY (id)
);

-- CREATE TABLE events.extra_life()
-- CREATE TABLE events.tiltify()
-- CREATE TABLE events.streamlabs()
-- CREATE TABLE events.streamelements()