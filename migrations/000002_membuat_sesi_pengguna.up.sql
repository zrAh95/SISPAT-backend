CREATE TABLE sesi_pengguna (
  id CHAR(64) NOT NULL,
  pengguna_id BIGINT UNSIGNED NOT NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  terakhir_aktif_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  kedaluwarsa_pada DATETIME NOT NULL,
  alamat_ip VARCHAR(45) NULL,
  user_agent VARCHAR(500) NULL,
  PRIMARY KEY (id),
  KEY idx_sesi_pengguna_pengguna (pengguna_id),
  KEY idx_sesi_pengguna_kedaluwarsa (kedaluwarsa_pada),
  CONSTRAINT fk_sesi_pengguna_pengguna FOREIGN KEY (pengguna_id) REFERENCES pengguna (id) ON DELETE CASCADE
) ENGINE=InnoDB;
