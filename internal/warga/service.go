package warga

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"sispat/backend/internal/auth"
	"sispat/backend/internal/models"
)

type CreateInput struct {
	NamaPengguna string `json:"nama_pengguna"`
	KataSandi    string `json:"kata_sandi"`
	Nama         string `json:"nama"`
	Alamat       string `json:"alamat"`
	NomorHP      string `json:"nomor_hp"`
	ReguID       uint   `json:"regu_id"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type Detail struct {
	models.Warga
	Regu    string `gorm:"column:regu" json:"regu"`
	Jabatan string `gorm:"column:jabatan" json:"jabatan"`
}

func (s *Service) List() ([]Detail, error) {
	var daftar []Detail
	err := s.db.Table("warga AS w").
		Select("w.*, r.nama AS regu, COALESCE(pr.jabatan, '') AS jabatan").
		Joins("JOIN anggota_regu AS ar ON ar.warga_id = w.id").
		Joins("JOIN regu AS r ON r.id = ar.regu_id").
		Joins("LEFT JOIN petugas_regu AS pr ON pr.warga_id = w.id").
		Order("w.nama ASC").Scan(&daftar).Error
	return daftar, err
}

func (s *Service) Get(id uint) (models.Warga, error) {
	var warga models.Warga
	if err := s.db.First(&warga, "id = ?", id).Error; err != nil {
		return models.Warga{}, err
	}
	return warga, nil
}

type UpdateInput struct {
	Nama    string `json:"nama"`
	Alamat  string `json:"alamat"`
	NomorHP string `json:"nomor_hp"`
	ReguID  uint   `json:"regu_id"`
}

func (s *Service) Update(id uint, input UpdateInput) (models.Warga, error) {
	input.Nama = strings.TrimSpace(input.Nama)
	input.Alamat = strings.TrimSpace(input.Alamat)
	input.NomorHP = strings.TrimSpace(input.NomorHP)
	if input.Nama == "" || input.Alamat == "" || input.NomorHP == "" || input.ReguID == 0 {
		return models.Warga{}, fmt.Errorf("data warga tidak lengkap")
	}

	var warga models.Warga
	if err := s.db.First(&warga, "id = ?", id).Error; err != nil {
		return models.Warga{}, err
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var regu models.Regu
		if err := tx.First(&regu, "id = ?", input.ReguID).Error; err != nil {
			return fmt.Errorf("regu tidak ditemukan")
		}
		if err := tx.Model(&warga).Updates(map[string]any{"nama": input.Nama, "alamat": input.Alamat, "nomor_hp": input.NomorHP}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.AnggotaRegu{}).Where("warga_id = ?", warga.ID).Update("regu_id", input.ReguID).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return models.Warga{}, err
	}
	return s.Get(id)
}

func (s *Service) HardDelete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var warga models.Warga
		if err := tx.First(&warga, "id = ?", id).Error; err != nil {
			return err
		}
		var historyCount int64
		if err := tx.Table("absensi").Where("warga_id = ?", id).Count(&historyCount).Error; err != nil {
			return err
		}
		if historyCount > 0 {
			return fmt.Errorf("warga tidak dapat dihapus karena memiliki riwayat absensi")
		}

		steps := []struct {
			query string
			args  []any
		}{
			{"DELETE FROM petugas_regu WHERE warga_id = ?", []any{id}},
			{"DELETE FROM anggota_regu WHERE warga_id = ?", []any{id}},
			{"DELETE FROM pengguna_peran WHERE pengguna_id = ?", []any{warga.PenggunaID}},
			{"DELETE FROM sesi_pengguna WHERE pengguna_id = ?", []any{warga.PenggunaID}},
			{"DELETE FROM warga WHERE id = ?", []any{id}},
			{"DELETE FROM pengguna WHERE id = ?", []any{warga.PenggunaID}},
		}
		for _, step := range steps {
			if err := tx.Exec(step.query, step.args...).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) SetActive(id uint, active bool) (models.Warga, error) {
	var warga models.Warga
	if err := s.db.First(&warga, "id = ?", id).Error; err != nil {
		return models.Warga{}, err
	}
	if err := s.db.Model(&warga).Update("aktif", active).Error; err != nil {
		return models.Warga{}, err
	}
	warga.Aktif = active
	return warga, nil
}

func (s *Service) Create(input CreateInput) (models.Warga, error) {
	input.NamaPengguna = strings.TrimSpace(input.NamaPengguna)
	input.Nama = strings.TrimSpace(input.Nama)
	input.Alamat = strings.TrimSpace(input.Alamat)
	input.NomorHP = strings.TrimSpace(input.NomorHP)
	if input.NamaPengguna == "" {
		input.NamaPengguna = strings.ToLower(strings.ReplaceAll(input.Nama, " ", "_"))
	}

	if err := validate(input); err != nil {
		return models.Warga{}, err
	}

	passwordHash, err := auth.HashPassword(input.KataSandi)
	if err != nil {
		return models.Warga{}, err
	}

	var warga models.Warga
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var pengguna models.Pengguna
		if err := tx.Where("nama_pengguna = ?", input.NamaPengguna).First(&pengguna).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		} else {
			return fmt.Errorf("nama warga sudah digunakan atau nama login bentrok")
		}

		var regu models.Regu
		if err := tx.First(&regu, "id = ?", input.ReguID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("regu tidak ditemukan")
			}
			return err
		}

		pengguna = models.Pengguna{
			NamaPengguna:  input.NamaPengguna,
			KataSandiHash: passwordHash,
			Aktif:         true,
		}
		if err := tx.Create(&pengguna).Error; err != nil {
			return err
		}

		warga = models.Warga{
			PenggunaID: pengguna.ID,
			Nama:       input.Nama,
			Alamat:     input.Alamat,
			NomorHP:    input.NomorHP,
			Aktif:      true,
		}
		if err := tx.Create(&warga).Error; err != nil {
			return err
		}

		anggota := models.AnggotaRegu{
			ReguID:  regu.ID,
			WargaID: warga.ID,
		}
		if err := tx.Create(&anggota).Error; err != nil {
			return err
		}

		var peran models.Peran
		if err := tx.Where("kode = ?", "ANGGOTA").First(&peran).Error; err != nil {
			return fmt.Errorf("peran ANGGOTA tidak ditemukan")
		}
		if err := tx.Create(&models.PenggunaPeran{
			PenggunaID: pengguna.ID,
			PeranID:    peran.ID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Warga{}, err
	}

	return s.Get(warga.ID)
}

func validate(input CreateInput) error {
	switch {
	case input.KataSandi == "":
		return fmt.Errorf("kata sandi wajib diisi")
	case input.Nama == "":
		return fmt.Errorf("nama warga wajib diisi")
	case input.Alamat == "":
		return fmt.Errorf("alamat wajib diisi")
	case input.NomorHP == "":
		return fmt.Errorf("nomor HP wajib diisi")
	case input.ReguID == 0:
		return fmt.Errorf("regu wajib dipilih")
	default:
		return nil
	}
}
