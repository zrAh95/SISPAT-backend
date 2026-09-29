package models

import (
	"time"

	"gorm.io/gorm"
)

type Peran struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Kode           string    `gorm:"size:30;uniqueIndex;not null" json:"kode"`
	Nama           string    `gorm:"size:80;not null" json:"nama"`
	DibuatPada     time.Time `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada time.Time `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
}

func (Peran) TableName() string { return "peran" }

type Pengguna struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	NamaPengguna      string         `gorm:"column:nama_pengguna;size:100;uniqueIndex;not null" json:"nama_pengguna"`
	KataSandiHash     string         `gorm:"column:kata_sandi_hash;size:255;not null" json:"-"`
	Aktif             bool           `gorm:"not null;default:true" json:"aktif"`
	TerakhirMasukPada *time.Time     `gorm:"column:terakhir_masuk_pada" json:"terakhir_masuk_pada"`
	DibuatPada        time.Time      `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada    time.Time      `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (Pengguna) TableName() string { return "pengguna" }

type Warga struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	PenggunaID     uint           `gorm:"column:pengguna_id;uniqueIndex;not null" json:"pengguna_id"`
	Nama           string         `gorm:"size:150;not null" json:"nama"`
	Alamat         string         `gorm:"type:text;not null" json:"alamat"`
	NomorHP        string         `gorm:"column:nomor_hp;size:30;not null" json:"nomor_hp"`
	Aktif          bool           `gorm:"not null;default:true" json:"aktif"`
	DibuatPada     time.Time      `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada time.Time      `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (Warga) TableName() string { return "warga" }

type PenggunaPeran struct {
	PenggunaID uint      `gorm:"primaryKey;column:pengguna_id" json:"pengguna_id"`
	PeranID    uint      `gorm:"primaryKey;column:peran_id" json:"peran_id"`
	DibuatPada time.Time `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
}

func (PenggunaPeran) TableName() string { return "pengguna_peran" }

type Regu struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Kode           string    `gorm:"size:10;uniqueIndex;not null" json:"kode"`
	Nama           string    `gorm:"size:80;not null" json:"nama"`
	DibuatPada     time.Time `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada time.Time `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
}

func (Regu) TableName() string { return "regu" }

type AnggotaRegu struct {
	ReguID         uint      `gorm:"primaryKey;column:regu_id" json:"regu_id"`
	WargaID        uint      `gorm:"primaryKey;column:warga_id" json:"warga_id"`
	DibuatPada     time.Time `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada time.Time `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
}

func (AnggotaRegu) TableName() string { return "anggota_regu" }

type PetugasRegu struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ReguID         uint      `gorm:"column:regu_id;uniqueIndex:idx_petugas_regu_per_regu;not null" json:"regu_id"`
	WargaID        uint      `gorm:"column:warga_id;uniqueIndex;not null" json:"warga_id"`
	Jabatan        string    `gorm:"size:20;uniqueIndex:idx_petugas_regu_per_regu;not null" json:"jabatan"`
	DibuatPada     time.Time `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada time.Time `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
}

func (PetugasRegu) TableName() string { return "petugas_regu" }

type JadwalRonda struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ReguID         uint      `gorm:"column:regu_id;not null" json:"regu_id"`
	TanggalRonda   time.Time `gorm:"column:tanggal_ronda;type:date;not null" json:"tanggal_ronda"`
	MulaiAt        time.Time `gorm:"column:mulai_at;not null" json:"mulai_at"`
	SelesaiAt      time.Time `gorm:"column:selesai_at;not null" json:"selesai_at"`
	Status         string    `gorm:"size:20;not null;default:TERJADWAL" json:"status"`
	DibuatPada     time.Time `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada time.Time `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
}

func (JadwalRonda) TableName() string { return "jadwal_ronda" }

type SesiRonda struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	JadwalID           uint       `gorm:"column:jadwal_id;uniqueIndex;not null" json:"jadwal_id"`
	Status             string     `gorm:"size:20;not null;default:TERJADWAL" json:"status"`
	DimulaiPada        *time.Time `gorm:"column:dimulai_pada" json:"dimulai_pada"`
	DiseselesaikanPada *time.Time `gorm:"column:diselesaikan_pada" json:"diselesaikan_pada"`
	DibuatPada         time.Time  `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada     time.Time  `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
}

func (SesiRonda) TableName() string { return "sesi_ronda" }

type Absensi struct {
	ID                         uint       `gorm:"primaryKey" json:"id"`
	SesiID                     uint       `gorm:"column:sesi_id;uniqueIndex:idx_absensi_sesi_warga;not null" json:"sesi_id"`
	WargaID                    uint       `gorm:"column:warga_id;uniqueIndex:idx_absensi_sesi_warga;not null" json:"warga_id"`
	Status                     string     `gorm:"size:20;not null" json:"status"`
	DiririmPada                time.Time  `gorm:"column:dikirim_pada;not null" json:"dikirim_pada"`
	DiverifikasiOlehPenggunaID *uint      `gorm:"column:diverifikasi_oleh_pengguna_id" json:"diverifikasi_oleh_pengguna_id"`
	DiverifikasiPada           *time.Time `gorm:"column:diverifikasi_pada" json:"diverifikasi_pada"`
	AlasanDitolak              string     `gorm:"column:alasan_ditolak;size:255" json:"alasan_ditolak"`
	DibuatPada                 time.Time  `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada             time.Time  `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
}

func (Absensi) TableName() string { return "absensi" }

type FotoAbsensi struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	AbsensiID      uint      `gorm:"column:absensi_id;uniqueIndex;not null" json:"absensi_id"`
	LokasiFile     string    `gorm:"column:lokasi_file;size:500;not null" json:"lokasi_file"`
	NamaAsliFile   string    `gorm:"column:nama_asli_file;size:255" json:"nama_asli_file"`
	TipeMime       string    `gorm:"column:tipe_mime;size:100;not null" json:"tipe_mime"`
	UkuranFile     uint64    `gorm:"column:ukuran_file;not null" json:"ukuran_file"`
	DibuatPada     time.Time `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada time.Time `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
}

func (FotoAbsensi) TableName() string { return "foto_absensi" }

type LaporanRonda struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	SesiID               uint      `gorm:"column:sesi_id;uniqueIndex;not null" json:"sesi_id"`
	StatusKegiatan       string    `gorm:"column:status_kegiatan;size:50;not null" json:"status_kegiatan"`
	RingkasanJimpitan    string    `gorm:"column:ringkasan_jimpitan;type:text" json:"ringkasan_jimpitan"`
	TotalJimpitan        float64   `gorm:"column:total_jimpitan;type:decimal(12,2);not null" json:"total_jimpitan"`
	Catatan              string    `gorm:"type:text" json:"catatan"`
	DibuatOlehPenggunaID uint      `gorm:"column:dibuat_oleh_pengguna_id;not null" json:"dibuat_oleh_pengguna_id"`
	DibuatPada           time.Time `gorm:"column:dibuat_pada;not null" json:"dibuat_pada"`
	DibuatSistemPada     time.Time `gorm:"column:waktu_dibuat;autoCreateTime" json:"-"`
	DiperbaruiSistemPada time.Time `gorm:"column:waktu_diubah;autoUpdateTime" json:"-"`
}

func (LaporanRonda) TableName() string { return "laporan_ronda" }

type BuktiLaporan struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	LaporanID      uint      `gorm:"column:laporan_id;uniqueIndex;not null" json:"laporan_id"`
	LokasiFile     string    `gorm:"column:lokasi_file;size:500;not null" json:"lokasi_file"`
	NamaAsliFile   string    `gorm:"column:nama_asli_file;size:255" json:"nama_asli_file"`
	TipeMime       string    `gorm:"column:tipe_mime;size:100;not null" json:"tipe_mime"`
	UkuranFile     uint64    `gorm:"column:ukuran_file;not null" json:"ukuran_file"`
	DibuatPada     time.Time `gorm:"column:dibuat_pada;autoCreateTime" json:"dibuat_pada"`
	DiperbaruiPada time.Time `gorm:"column:diperbarui_pada;autoUpdateTime" json:"diperbarui_pada"`
}

func (BuktiLaporan) TableName() string { return "bukti_laporan" }
