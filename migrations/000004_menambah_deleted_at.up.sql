ALTER TABLE pengguna
  ADD COLUMN deleted_at DATETIME NULL AFTER diperbarui_pada;

ALTER TABLE warga
  ADD COLUMN deleted_at DATETIME NULL AFTER diperbarui_pada;

CREATE INDEX idx_pengguna_deleted_at ON pengguna (deleted_at);
CREATE INDEX idx_warga_deleted_at ON warga (deleted_at);
