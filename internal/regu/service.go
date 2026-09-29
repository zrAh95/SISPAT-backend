package regu

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"sispat/backend/internal/models"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type Member struct {
	WargaID uint   `json:"warga_id"`
	Nama    string `json:"nama"`
	Alamat  string `json:"alamat"`
	NomorHP string `json:"nomor_hp"`
	Aktif   bool   `json:"aktif"`
}

type Officer struct {
	WargaID uint   `json:"warga_id"`
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
}

func (s *Service) List() ([]models.Regu, error) {
	var daftar []models.Regu
	err := s.db.Order("kode ASC").Find(&daftar).Error
	return daftar, err
}

func (s *Service) Get(id uint) (models.Regu, error) {
	var data models.Regu
	if err := s.db.First(&data, "id = ?", id).Error; err != nil {
		return models.Regu{}, err
	}
	return data, nil
}

func (s *Service) Members(groupID uint) ([]Member, error) {
	var daftar []Member
	err := s.db.Table("anggota_regu AS ar").
		Select("w.id AS warga_id, w.nama, w.alamat, w.nomor_hp, w.aktif").
		Joins("JOIN warga AS w ON w.id = ar.warga_id").
		Where("ar.regu_id = ?", groupID).
		Order("w.nama ASC").
		Scan(&daftar).Error
	return daftar, err
}

func (s *Service) AssignOfficer(groupID, wargaID uint, jabatan string) error {
	if jabatan != "DANRU" && jabatan != "WAKIL_DANRU" {
		return fmt.Errorf("jabatan tidak valid")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var member models.AnggotaRegu
		if err := tx.First(&member, "regu_id = ? AND warga_id = ?", groupID, wargaID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("warga harus menjadi anggota regu terlebih dahulu")
			}
			return err
		}

		if err := tx.Where("regu_id = ? AND jabatan = ?", groupID, jabatan).Delete(&models.PetugasRegu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("warga_id = ?", wargaID).Delete(&models.PetugasRegu{}).Error; err != nil {
			return err
		}

		return tx.Create(&models.PetugasRegu{
			ReguID:  groupID,
			WargaID: wargaID,
			Jabatan: jabatan,
		}).Error
	})
}

func (s *Service) Officers(groupID uint) ([]Officer, error) {
	var daftar []Officer
	err := s.db.Table("petugas_regu AS pr").
		Select("w.id AS warga_id, w.nama, pr.jabatan").
		Joins("JOIN warga AS w ON w.id = pr.warga_id").
		Where("pr.regu_id = ?", groupID).
		Order("pr.jabatan ASC").
		Scan(&daftar).Error
	return daftar, err
}
