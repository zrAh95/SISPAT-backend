ALTER TABLE sesi_pengguna
  ADD COLUMN csrf_token CHAR(64) NOT NULL AFTER id;
