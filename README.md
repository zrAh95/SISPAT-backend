# SISPAT Backend

API untuk Sistem Informasi Ronda Khusus RT Empat.

## Stack

```text
Go
chi           HTTP router
GORM          ORM
golang-migrate  Migrasi database
MySQL 8.4     Database
Argon2id      Hashing password
```

## Menjalankan

Buat file `.env` dengan menyalin `.env.example`:

```bash
cp .env.example .env
```

Isi `DB_PASSWORD` sesuai password MySQL lokal, lalu:

```bash
set -a
source .env
set +a
go run ./cmd/api
```

API berjalan di `http://localhost:8080`.

Cek kesehatan server:

```bash
curl http://localhost:8080/api/health
```

## Migrasi Database

```bash
go run ./cmd/migrate up
```

Status migrasi:

```bash
go run ./cmd/migrate version
```

Rollback satu langkah:

```bash
go run ./cmd/migrate down 1
```

## Seed Akun

Seed akun awal Pak RT dan Koordinator. Password dibaca lewat input tersembunyi dan hanya disimpan sebagai hash.

```bash
go run ./cmd/seed
```

Seed akun testing untuk kelima role sekaligus:

```bash
go run ./cmd/seed-testing
```

## Test

```bash
go test ./...
```

## Peran

```text
PAK_RT          Membuat jadwal, menetapkan Koordinator, monitoring
KOORDINATOR     Kelola warga, kelola regu, atur Danru dan Wakil
DANRU           Mulai sesi, verifikasi absensi, kirim laporan
WAKIL_DANRU     Sama seperti Danru
ANGGOTA         Absensi pribadi
```

## Struktur

```text
cmd/api           Server utama dan routing
cmd/migrate       Runner migrasi
cmd/seed          Seed akun awal
cmd/seed-testing  Seed akun testing
internal/auth     Session, CSRF, RBAC, Argon2id
internal/warga    CRUD warga
internal/regu     Anggota regu dan petugas
internal/jadwal   Jadwal ronda
internal/absensi  Absensi dan sesi ronda
internal/verifikasi  Verifikasi kehadiran
internal/laporan  Laporan ronda
internal/koordinator  Penetapan role Koordinator
internal/dashboard Dashboard dan rekap
internal/http     Handler HTTP
migrations        File migrasi SQL
```
