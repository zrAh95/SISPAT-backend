package main

import (
	"bufio"
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

type seedPengguna struct {
	namaPengguna string
	kataSandi    string
	kodePeran    string
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

	fmt.Println("Seed akun awal SISPAT")
	fmt.Println("Password tidak akan ditampilkan dan tidak disimpan sebagai plaintext.")

	pakRT := bacaPengguna("Pak RT", "PAK_RT")
	koordinator := bacaPengguna("Koordinator", "KOORDINATOR")

	if err := seedPenggunaAwal(db, []seedPengguna{pakRT, koordinator}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Selesai. Akun awal berhasil disimpan.")
}

func bacaPengguna(label, kodePeran string) seedPengguna {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Username %s: ", label)
	namaPengguna, _ := reader.ReadString('\n')
	namaPengguna = strings.TrimSpace(namaPengguna)

	fmt.Printf("Password %s: ", label)
	kataSandi, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		log.Fatalf("gagal membaca password %s: %v", label, err)
	}

	return seedPengguna{
		namaPengguna: namaPengguna,
		kataSandi:    string(kataSandi),
		kodePeran:    kodePeran,
	}
}

func seedPenggunaAwal(db *gorm.DB, seeds []seedPengguna) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, seed := range seeds {
			if seed.namaPengguna == "" {
				return fmt.Errorf("username %s tidak boleh kosong", seed.kodePeran)
			}

			passwordHash, err := auth.HashPassword(seed.kataSandi)
			if err != nil {
				return fmt.Errorf("hash password %s: %w", seed.namaPengguna, err)
			}

			var pengguna models.Pengguna
			result := tx.Where("nama_pengguna = ?", seed.namaPengguna).Find(&pengguna)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				pengguna = models.Pengguna{
					NamaPengguna:  seed.namaPengguna,
					KataSandiHash: passwordHash,
					Aktif:         true,
				}
				if err := tx.Create(&pengguna).Error; err != nil {
					return err
				}
			} else {
				fmt.Printf("Pengguna %s sudah ada, dilewati.\n", seed.namaPengguna)
			}

			var peran models.Peran
			if err := tx.Where("kode = ?", seed.kodePeran).First(&peran).Error; err != nil {
				return fmt.Errorf("peran %s tidak ditemukan: %w", seed.kodePeran, err)
			}

			var assignment models.PenggunaPeran
			assignmentResult := tx.Where("pengguna_id = ? AND peran_id = ?", pengguna.ID, peran.ID).Find(&assignment)
			if assignmentResult.Error != nil {
				return assignmentResult.Error
			}
			if assignmentResult.RowsAffected == 0 {
				assignment = models.PenggunaPeran{
					PenggunaID: pengguna.ID,
					PeranID:    peran.ID,
				}
				if err := tx.Create(&assignment).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
