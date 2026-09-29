package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/term"
	"gorm.io/gorm"

	"sispat/backend/internal/auth"
	"sispat/backend/internal/config"
	"sispat/backend/internal/database"
	"sispat/backend/internal/models"
)

type akun struct {
	username string
	password string
	nama     string
	alamat   string
	nomorHP  string
	role     string
	office   string
	reguID   uint
}

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("gagal mengambil koneksi database: %v", err)
	}
	defer sqlDB.Close()

	akunTesting := []akun{
		{username: baca("Username Pak RT", "pak_rt"), password: bacaPassword("Pak RT"), role: "PAK_RT"},
		{username: baca("Username Koordinator", "koordinator"), password: bacaPassword("Koordinator"), role: "KOORDINATOR"},
		{
			username: baca("Username Danru", "danru"), password: bacaPassword("Danru"),
			nama: baca("Nama Danru", "Danru Testing"), alamat: baca("Alamat Danru", "Alamat Danru Testing"),
			nomorHP: baca("Nomor HP Danru", "081200000001"), role: "DANRU", office: "DANRU", reguID: 1,
		},
		{
			username: baca("Username Wakil Danru", "wakil"), password: bacaPassword("Wakil Danru"),
			nama: baca("Nama Wakil Danru", "Wakil Danru Testing"), alamat: baca("Alamat Wakil Danru", "Alamat Wakil Testing"),
			nomorHP: baca("Nomor HP Wakil Danru", "081200000003"), role: "WAKIL_DANRU", office: "WAKIL_DANRU", reguID: 1,
		},
		{
			username: baca("Username Anggota", "anggota"), password: bacaPassword("Anggota"),
			nama: baca("Nama Anggota", "Anggota Testing"), alamat: baca("Alamat Anggota", "Alamat Anggota Testing"),
			nomorHP: baca("Nomor HP Anggota", "081200000002"), role: "ANGGOTA", office: "", reguID: 1,
		},
	}

	for _, data := range akunTesting {
		if err := seedAkun(db, data); err != nil {
			log.Fatalf("gagal seed %s: %v", data.username, err)
		}
	}
	fmt.Println("Seed akun testing selesai.")
}

func seedAkun(db *gorm.DB, data akun) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var pengguna models.Pengguna
		result := tx.Where("nama_pengguna = ?", data.username).Find(&pengguna)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			fmt.Printf("Pengguna %s sudah ada, dilewati.\n", data.username)
		} else {
			hash, err := auth.HashPassword(data.password)
			if err != nil {
				return err
			}
			pengguna = models.Pengguna{NamaPengguna: data.username, KataSandiHash: hash, Aktif: true}
			if err := tx.Create(&pengguna).Error; err != nil {
				return err
			}
		}

		var role models.Peran
		if err := tx.Where("kode = ?", data.role).First(&role).Error; err != nil {
			return err
		}
		if err := tx.Where("pengguna_id = ? AND peran_id = ?", pengguna.ID, role.ID).First(&models.PenggunaPeran{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&models.PenggunaPeran{PenggunaID: pengguna.ID, PeranID: role.ID}).Error; err != nil {
				return err
			}
		}

		if data.role == "PAK_RT" || data.role == "KOORDINATOR" {
			return nil
		}

		var warga models.Warga
		result = tx.Where("pengguna_id = ?", pengguna.ID).Find(&warga)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			warga = models.Warga{PenggunaID: pengguna.ID, Nama: data.nama, Alamat: data.alamat, NomorHP: data.nomorHP, Aktif: true}
			if err := tx.Create(&warga).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("regu_id = ? AND warga_id = ?", data.reguID, warga.ID).First(&models.AnggotaRegu{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&models.AnggotaRegu{ReguID: data.reguID, WargaID: warga.ID}).Error; err != nil {
				return err
			}
		}
		if data.office != "" {
			if err := tx.Where("regu_id = ? AND jabatan = ?", data.reguID, data.office).Delete(&models.PetugasRegu{}).Error; err != nil {
				return err
			}
			if err := tx.Where("warga_id = ?", warga.ID).Delete(&models.PetugasRegu{}).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.PetugasRegu{ReguID: data.reguID, WargaID: warga.ID, Jabatan: data.office}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func reader() *bufio.Reader { return bufio.NewReader(os.Stdin) }

func baca(label, defaultValue string) string {
	fmt.Printf("%s [%s]: ", label, defaultValue)
	value, _ := reader().ReadString('\n')
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue
	}
	return value
}

func bacaPassword(label string) string {
	fmt.Printf("Password %s: ", label)
	value, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		log.Fatal(err)
	}
	return string(value)
}
