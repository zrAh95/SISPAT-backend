CREATE TABLE peran (
  id TINYINT UNSIGNED NOT NULL AUTO_INCREMENT,
  kode VARCHAR(30) NOT NULL,
  nama VARCHAR(80) NOT NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_peran_kode (kode)
) ENGINE=InnoDB;

CREATE TABLE pengguna (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  nama_pengguna VARCHAR(100) NOT NULL,
  kata_sandi_hash VARCHAR(255) NOT NULL,
  aktif BOOLEAN NOT NULL DEFAULT TRUE,
  terakhir_masuk_pada DATETIME NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_pengguna_nama (nama_pengguna),
  KEY idx_pengguna_aktif (aktif)
) ENGINE=InnoDB;

CREATE TABLE warga (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  pengguna_id BIGINT UNSIGNED NOT NULL,
  nama VARCHAR(150) NOT NULL,
  alamat TEXT NOT NULL,
  nomor_hp VARCHAR(30) NOT NULL,
  aktif BOOLEAN NOT NULL DEFAULT TRUE,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_warga_pengguna (pengguna_id),
  KEY idx_warga_aktif (aktif),
  CONSTRAINT fk_warga_pengguna FOREIGN KEY (pengguna_id) REFERENCES pengguna (id) ON DELETE RESTRICT
) ENGINE=InnoDB;

CREATE TABLE pengguna_peran (
  pengguna_id BIGINT UNSIGNED NOT NULL,
  peran_id TINYINT UNSIGNED NOT NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (pengguna_id, peran_id),
  KEY idx_pengguna_peran_peran (peran_id),
  CONSTRAINT fk_pengguna_peran_pengguna FOREIGN KEY (pengguna_id) REFERENCES pengguna (id) ON DELETE CASCADE,
  CONSTRAINT fk_pengguna_peran_peran FOREIGN KEY (peran_id) REFERENCES peran (id) ON DELETE RESTRICT
) ENGINE=InnoDB;

CREATE TABLE regu (
  id TINYINT UNSIGNED NOT NULL AUTO_INCREMENT,
  kode VARCHAR(10) NOT NULL,
  nama VARCHAR(80) NOT NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_regu_kode (kode)
) ENGINE=InnoDB;

CREATE TABLE anggota_regu (
  regu_id TINYINT UNSIGNED NOT NULL,
  warga_id BIGINT UNSIGNED NOT NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (regu_id, warga_id),
  UNIQUE KEY unik_anggota_regu_warga (warga_id),
  KEY idx_anggota_regu_regu (regu_id),
  CONSTRAINT fk_anggota_regu_regu FOREIGN KEY (regu_id) REFERENCES regu (id) ON DELETE RESTRICT,
  CONSTRAINT fk_anggota_regu_warga FOREIGN KEY (warga_id) REFERENCES warga (id) ON DELETE RESTRICT
) ENGINE=InnoDB;

CREATE TABLE petugas_regu (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  regu_id TINYINT UNSIGNED NOT NULL,
  warga_id BIGINT UNSIGNED NOT NULL,
  jabatan ENUM ('DANRU', 'WAKIL_DANRU') NOT NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_petugas_regu_per_regu (regu_id, jabatan),
  UNIQUE KEY unik_petugas_regu_warga (warga_id),
  KEY idx_petugas_regu_warga (warga_id),
  CONSTRAINT fk_petugas_regu_anggota FOREIGN KEY (regu_id, warga_id) REFERENCES anggota_regu (regu_id, warga_id) ON DELETE RESTRICT
) ENGINE=InnoDB;

CREATE TABLE jadwal_ronda (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  regu_id TINYINT UNSIGNED NOT NULL,
  tanggal_ronda DATE NOT NULL,
  mulai_at DATETIME NOT NULL,
  selesai_at DATETIME NOT NULL,
  status ENUM ('TERJADWAL', 'SELESAI', 'DIBATALKAN') NOT NULL DEFAULT 'TERJADWAL',
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_jadwal_ronda_regu_tanggal (regu_id, tanggal_ronda),
  KEY idx_jadwal_ronda_tanggal (tanggal_ronda),
  CONSTRAINT fk_jadwal_ronda_regu FOREIGN KEY (regu_id) REFERENCES regu (id) ON DELETE RESTRICT,
  CONSTRAINT cek_waktu_jadwal CHECK (selesai_at > mulai_at)
) ENGINE=InnoDB;

CREATE TABLE sesi_ronda (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  jadwal_id BIGINT UNSIGNED NOT NULL,
  status ENUM ('TERJADWAL', 'BERLANGSUNG', 'SELESAI', 'DIBATALKAN') NOT NULL DEFAULT 'TERJADWAL',
  dimulai_pada DATETIME NULL,
  diselesaikan_pada DATETIME NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_sesi_ronda_jadwal (jadwal_id),
  KEY idx_sesi_ronda_status (status),
  CONSTRAINT fk_sesi_ronda_jadwal FOREIGN KEY (jadwal_id) REFERENCES jadwal_ronda (id) ON DELETE RESTRICT
) ENGINE=InnoDB;

CREATE TABLE absensi (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  sesi_id BIGINT UNSIGNED NOT NULL,
  warga_id BIGINT UNSIGNED NOT NULL,
  status ENUM ('MENUNGGU', 'TERVERIFIKASI', 'DITOLAK') NOT NULL,
  dikirim_pada DATETIME NOT NULL,
  diverifikasi_oleh_pengguna_id BIGINT UNSIGNED NULL,
  diverifikasi_pada DATETIME NULL,
  alasan_ditolak VARCHAR(255) NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_absensi_sesi_warga (sesi_id, warga_id),
  KEY idx_absensi_sesi_status (sesi_id, status),
  KEY idx_absensi_warga (warga_id),
  KEY idx_absensi_diverifikasi_oleh (diverifikasi_oleh_pengguna_id),
  CONSTRAINT fk_absensi_sesi FOREIGN KEY (sesi_id) REFERENCES sesi_ronda (id) ON DELETE RESTRICT,
  CONSTRAINT fk_absensi_warga FOREIGN KEY (warga_id) REFERENCES warga (id) ON DELETE RESTRICT,
  CONSTRAINT fk_absensi_diverifikasi_oleh FOREIGN KEY (diverifikasi_oleh_pengguna_id) REFERENCES pengguna (id) ON DELETE SET NULL
) ENGINE=InnoDB;

CREATE TABLE foto_absensi (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  absensi_id BIGINT UNSIGNED NOT NULL,
  lokasi_file VARCHAR(500) NOT NULL,
  nama_asli_file VARCHAR(255) NULL,
  tipe_mime VARCHAR(100) NOT NULL,
  ukuran_file BIGINT UNSIGNED NOT NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_foto_absensi (absensi_id),
  CONSTRAINT fk_foto_absensi_absensi FOREIGN KEY (absensi_id) REFERENCES absensi (id) ON DELETE CASCADE,
  CONSTRAINT cek_ukuran_foto_absensi CHECK (ukuran_file > 0)
) ENGINE=InnoDB;

CREATE TABLE laporan_ronda (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  sesi_id BIGINT UNSIGNED NOT NULL,
  status_kegiatan VARCHAR(50) NOT NULL,
  ringkasan_jimpitan TEXT NULL,
  total_jimpitan DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  catatan TEXT NULL,
  dibuat_oleh_pengguna_id BIGINT UNSIGNED NOT NULL,
  dibuat_pada DATETIME NOT NULL,
  waktu_dibuat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  waktu_diubah TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_laporan_sesi (sesi_id),
  KEY idx_laporan_dibuat_oleh (dibuat_oleh_pengguna_id),
  CONSTRAINT fk_laporan_sesi FOREIGN KEY (sesi_id) REFERENCES sesi_ronda (id) ON DELETE RESTRICT,
  CONSTRAINT fk_laporan_dibuat_oleh FOREIGN KEY (dibuat_oleh_pengguna_id) REFERENCES pengguna (id) ON DELETE RESTRICT,
  CONSTRAINT cek_total_jimpitan CHECK (total_jimpitan >= 0)
) ENGINE=InnoDB;

CREATE TABLE bukti_laporan (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  laporan_id BIGINT UNSIGNED NOT NULL,
  lokasi_file VARCHAR(500) NOT NULL,
  nama_asli_file VARCHAR(255) NULL,
  tipe_mime VARCHAR(100) NOT NULL,
  ukuran_file BIGINT UNSIGNED NOT NULL,
  dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diperbarui_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY unik_bukti_laporan (laporan_id),
  CONSTRAINT fk_bukti_laporan_laporan FOREIGN KEY (laporan_id) REFERENCES laporan_ronda (id) ON DELETE CASCADE,
  CONSTRAINT cek_ukuran_bukti_laporan CHECK (ukuran_file > 0)
) ENGINE=InnoDB;

INSERT INTO peran (kode, nama) VALUES
  ('PAK_RT', 'Pak RT'),
  ('KOORDINATOR', 'Koordinator Keamanan/Ronda'),
  ('DANRU', 'Danru'),
  ('WAKIL_DANRU', 'Wakil Danru'),
  ('ANGGOTA', 'Anggota Regu');

INSERT INTO regu (kode, nama) VALUES
  ('A', 'Regu A'),
  ('B', 'Regu B'),
  ('C', 'Regu C'),
  ('D', 'Regu D');
