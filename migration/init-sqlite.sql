CREATE TABLE oauth_token (
  client_id TEXT NOT NULL,
  access_token TEXT NOT NULL,
  refresh_token TEXT NOT NULL,
  scope TEXT,
  token_type TEXT NOT NULL,
  expires_at TEXT NOT NULL,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,

  PRIMARY KEY (client_id)
);

CREATE TABLE twitch_viewer(
  id TEXT NOT NULL,
  username TEXT NOT NULL,

  PRIMARY KEY (id)
);

CREATE TABLE events_twitch (
  id TEXT NOT NULL,
  type TEXT NOT NULL,
  amount TEXT,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,

  PRIMARY KEY (id)
);

-- CREATE TABLE events_extra_life()
-- CREATE TABLE events_tiltify()
-- CREATE TABLE events_streamlabs()
-- CREATE TABLE events_streamelements()