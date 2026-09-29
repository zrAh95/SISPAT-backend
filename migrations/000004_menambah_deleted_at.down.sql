DROP INDEX idx_warga_deleted_at ON warga;
DROP INDEX idx_pengguna_deleted_at ON pengguna;
ALTER TABLE warga DROP COLUMN deleted_at;
ALTER TABLE pengguna DROP COLUMN deleted_at;
