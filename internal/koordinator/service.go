package koordinator

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"sispat/backend/internal/models"
)

const KodePeran = "KOORDINATOR"

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type Calon struct {
	WargaID uint   `json:"warga_id"`
	Nama    string `json:"nama"`
	Regu    string `json:"regu"`
	NomorHP string `json:"nomor_hp"`
	Sudah   bool   `json:"sudah_koordinator"`
	Peran   string `json:"peran_utama"`
}

// ListCalon menampilkan warga beserta status Koordinator untuk dipilih Pak RT.
func (s *Service) ListCalon() ([]Calon, error) {
	var daftar []Calon
	err := s.db.Table("warga AS w").
		Select(`w.id AS warga_id, w.nama, w.nomor_hp, COALESCE(r.nama, '') AS regu,
			COALESCE((SELECT p.kode FROM pengguna_peran pp JOIN peran p ON p.id = pp.peran_id WHERE pp.pengguna_id = w.pengguna_id ORDER BY p.id LIMIT 1), '') AS peran,
			EXISTS(SELECT 1 FROM pengguna_peran pp2 JOIN peran p2 ON p2.id = pp2.peran_id WHERE pp2.pengguna_id = w.pengguna_id AND p2.kode = ?) AS sudah`, KodePeran).
		Where("w.aktif = TRUE").
		Joins("LEFT JOIN anggota_regu AS ar ON ar.warga_id = w.id").
		Joins("LEFT JOIN regu AS r ON r.id = ar.regu_id").
		Order("w.nama ASC").
		Scan(&daftar).Error
	if daftar == nil {
		daftar = make([]Calon, 0)
	}
	return daftar, err
}

func (s *Service) Assign(wargaID uint) error {
	if wargaID == 0 {
		return fmt.Errorf("warga wajib dipilih")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var warga models.Warga
		if err := tx.First(&warga, "id = ? AND aktif = TRUE", wargaID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("warga tidak ditemukan atau sudah tidak aktif")
			}
			return err
		}

		var peran models.Peran
		if err := tx.Where("kode = ?", KodePeran).First(&peran).Error; err != nil {
			return fmt.Errorf("peran koordinator tidak ditemukan")
		}

		var existing models.PenggunaPeran
		err := tx.Where("pengguna_id = ? AND peran_id = ?", warga.PenggunaID, peran.ID).First(&existing).Error
		if err == nil {
			return fmt.Errorf("warga ini sudah menjadi koordinator")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return tx.Create(&models.PenggunaPeran{PenggunaID: warga.PenggunaID, PeranID: peran.ID}).Error
	})
}
