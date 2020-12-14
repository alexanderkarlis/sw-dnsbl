CREATE TABLE IF NOT EXISTS ip_details (
  ip_address TEXT PRIMARY KEY NOT NULL, 
  uuid TEXT NOT NULL, 
  response_code text,
  created_at TEXT, 
  updated_at TEXT
);
