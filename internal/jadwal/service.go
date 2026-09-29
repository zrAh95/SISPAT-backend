package jadwal

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"sispat/backend/internal/models"
)

type CreateInput struct {
	ReguID       uint   `json:"regu_id"`
	TanggalRonda string `json:"tanggal_ronda"`
	MulaiAt      string `json:"mulai_at"`
	SelesaiAt    string `json:"selesai_at"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) List() ([]models.JadwalRonda, error) {
	var daftar []models.JadwalRonda
	err := s.db.Order("tanggal_ronda DESC, mulai_at DESC").Find(&daftar).Error
	return daftar, err
}

func (s *Service) Get(id uint) (models.JadwalRonda, error) {
	var data models.JadwalRonda
	if err := s.db.First(&data, "id = ?", id).Error; err != nil {
		return models.JadwalRonda{}, err
	}
	return data, nil
}

func (s *Service) Create(input CreateInput) (models.JadwalRonda, error) {
	if input.ReguID == 0 {
		return models.JadwalRonda{}, fmt.Errorf("regu wajib dipilih")
	}
	tanggal, err := time.Parse("2006-01-02", input.TanggalRonda)
	if err != nil {
		return models.JadwalRonda{}, fmt.Errorf("tanggal_ronda tidak valid")
	}
	mulai, err := parseWaktu(input.MulaiAt)
	if err != nil {
		return models.JadwalRonda{}, fmt.Errorf("mulai_at tidak valid")
	}
	selesai, err := parseWaktu(input.SelesaiAt)
	if err != nil {
		return models.JadwalRonda{}, fmt.Errorf("selesai_at tidak valid")
	}
	if !selesai.After(mulai) {
		return models.JadwalRonda{}, fmt.Errorf("selesai_at harus lebih besar dari mulai_at")
	}

	var jadwal models.JadwalRonda
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var regu models.Regu
		if err := tx.First(&regu, "id = ?", input.ReguID).Error; err != nil {
			return fmt.Errorf("regu tidak ditemukan")
		}
		jadwal = models.JadwalRonda{
			ReguID:       input.ReguID,
			TanggalRonda: tanggal,
			MulaiAt:      mulai,
			SelesaiAt:    selesai,
			Status:       "TERJADWAL",
		}
		if err := tx.Create(&jadwal).Error; err != nil {
			return err
		}
		return tx.Create(&models.SesiRonda{
			JadwalID: jadwal.ID,
			Status:   "TERJADWAL",
		}).Error
	})
	if err != nil {
		return models.JadwalRonda{}, err
	}
	return s.Get(jadwal.ID)
}

func parseWaktu(value string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05", value)
}
