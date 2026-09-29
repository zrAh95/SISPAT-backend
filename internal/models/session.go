package models

import "time"

type SesiPengguna struct {
	ID                string    `gorm:"primaryKey;size:64" json:"-"`
	CSRFToken         string    `gorm:"column:csrf_token;size:64;not null" json:"-"`
	PenggunaID        uint      `gorm:"column:pengguna_id;not null" json:"-"`
	DibuatPada        time.Time `gorm:"column:dibuat_pada" json:"-"`
	TerakhirAktifPada time.Time `gorm:"column:terakhir_aktif_pada" json:"-"`
	KedaluwarsaPada   time.Time `gorm:"column:kedaluwarsa_pada;not null" json:"-"`
	AlamatIP          string    `gorm:"column:alamat_ip;size:45" json:"-"`
	UserAgent         string    `gorm:"column:user_agent;size:500" json:"-"`
}

func (SesiPengguna) TableName() string { return "sesi_pengguna" }
