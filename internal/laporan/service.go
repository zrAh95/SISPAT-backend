package laporan

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"sispat/backend/internal/config"
	"sispat/backend/internal/models"
)

type Service struct {
	db  *gorm.DB
	cfg config.Config
}

func NewService(db *gorm.DB, cfg config.Config) *Service { return &Service{db: db, cfg: cfg} }

type CreateInput struct {
	StatusKegiatan    string  `json:"status_kegiatan"`
	RingkasanJimpitan string  `json:"ringkasan_jimpitan"`
	TotalJimpitan     float64 `json:"total_jimpitan"`
	Catatan           string  `json:"catatan"`
}

type View struct {
	ID                   uint      `json:"id"`
	SesiID               uint      `json:"sesi_id"`
	StatusKegiatan       string    `json:"status_kegiatan"`
	RingkasanJimpitan    string    `json:"ringkasan_jimpitan"`
	TotalJimpitan        float64   `json:"total_jimpitan"`
	Catatan              string    `json:"catatan"`
	DibuatOlehPenggunaID uint      `json:"dibuat_oleh_pengguna_id"`
	DibuatPada           time.Time `json:"dibuat_pada"`
}

func (s *Service) Create(sessionID, userID uint, roles []string, input CreateInput) (View, error) {
	if strings.TrimSpace(input.StatusKegiatan) == "" {
		return View{}, fmt.Errorf("status kegiatan wajib diisi")
	}
	if input.TotalJimpitan < 0 {
		return View{}, fmt.Errorf("total jimpitan tidak boleh negatif")
	}
	if !hasRole(roles, "DANRU") && !hasRole(roles, "WAKIL_DANRU") {
		return View{}, fmt.Errorf("hanya Danru atau Wakil Danru yang dapat membuat laporan")
	}
	groupID, err := s.sessionGroup(sessionID)
	if err != nil {
		return View{}, err
	}
	if err := s.authorizeOfficer(userID, groupID, roles, sessionID); err != nil {
		return View{}, err
	}
	var report models.LaporanRonda
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&report, "sesi_id = ?", sessionID).Error; err == nil {
			return fmt.Errorf("laporan sesi ini sudah ada")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		now := time.Now()
		report = models.LaporanRonda{
			SesiID:               sessionID,
			StatusKegiatan:       input.StatusKegiatan,
			RingkasanJimpitan:    input.RingkasanJimpitan,
			TotalJimpitan:        input.TotalJimpitan,
			Catatan:              input.Catatan,
			DibuatOlehPenggunaID: userID,
			DibuatPada:           now,
		}
		if err := tx.Create(&report).Error; err != nil {
			return err
		}
		return tx.Model(&models.SesiRonda{}).Where("id = ?", sessionID).Updates(map[string]any{"status": "SELESAI", "diselesaikan_pada": now}).Error
	})
	if err != nil {
		return View{}, err
	}
	return s.Get(report.ID)
}

func (s *Service) Get(id uint) (View, error) {
	var report models.LaporanRonda
	if err := s.db.First(&report, "id = ?", id).Error; err != nil {
		return View{}, err
	}
	return View{ID: report.ID, SesiID: report.SesiID, StatusKegiatan: report.StatusKegiatan, RingkasanJimpitan: report.RingkasanJimpitan, TotalJimpitan: report.TotalJimpitan, Catatan: report.Catatan, DibuatOlehPenggunaID: report.DibuatOlehPenggunaID, DibuatPada: report.DibuatPada}, nil
}

func (s *Service) AddEvidence(reportID, userID uint, roles []string, file multipart.File, header *multipart.FileHeader) error {
	if header.Size <= 0 || header.Size > s.cfg.MaxUploadBytes {
		return fmt.Errorf("ukuran file maksimal 5 MB")
	}
	mimeType, extension, err := validatePhoto(file)
	if err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	var report models.LaporanRonda
	if err := s.db.First(&report, "id = ?", reportID).Error; err != nil {
		return fmt.Errorf("laporan tidak ditemukan")
	}
	groupID, err := s.sessionGroup(report.SesiID)
	if err != nil {
		return err
	}
	if err := s.authorizeOfficer(userID, groupID, roles, report.SesiID); err != nil {
		return err
	}
	var old models.BuktiLaporan
	_ = s.db.First(&old, "laporan_id = ?", reportID).Error
	dir := filepath.Join(s.cfg.UploadDir, "laporan")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("gagal menyiapkan folder evidence")
	}
	path := filepath.Join(dir, randomName(extension))
	destination, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return fmt.Errorf("gagal menyimpan evidence")
	}
	if _, err := io.Copy(destination, file); err != nil {
		_ = destination.Close()
		_ = os.Remove(path)
		return fmt.Errorf("gagal menyimpan evidence")
	}
	_ = destination.Close()
	if old.LokasiFile != "" {
		_ = os.Remove(old.LokasiFile)
	}
	if old.ID == 0 {
		return s.db.Create(&models.BuktiLaporan{LaporanID: reportID, LokasiFile: path, NamaAsliFile: header.Filename, TipeMime: mimeType, UkuranFile: uint64(header.Size)}).Error
	}
	return s.db.Model(&models.BuktiLaporan{}).Where("id = ?", old.ID).Updates(map[string]any{"lokasi_file": path, "nama_asli_file": header.Filename, "tipe_mime": mimeType, "ukuran_file": header.Size}).Error
}

func (s *Service) sessionGroup(sessionID uint) (uint, error) {
	var groupID uint
	if err := s.db.Table("sesi_ronda AS s").Select("j.regu_id").Joins("JOIN jadwal_ronda AS j ON j.id = s.jadwal_id").Where("s.id = ?", sessionID).Scan(&groupID).Error; err != nil || groupID == 0 {
		return 0, fmt.Errorf("sesi ronda tidak ditemukan")
	}
	return groupID, nil
}

func (s *Service) authorizeOfficer(userID, groupID uint, roles []string, sessionID uint) error {
	var count int64
	if err := s.db.Table("petugas_regu AS pr").Joins("JOIN warga AS w ON w.id = pr.warga_id").Where("w.pengguna_id = ? AND pr.regu_id = ?", userID, groupID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("Anda bukan petugas regu ini")
	}
	if !hasRole(roles, "WAKIL_DANRU") || hasRole(roles, "DANRU") {
		return nil
	}
	var danruID uint
	_ = s.db.Table("petugas_regu AS pr").Joins("JOIN warga AS w ON w.id = pr.warga_id").Where("pr.regu_id = ? AND pr.jabatan = 'DANRU'", groupID).Select("w.id").Scan(&danruID).Error
	var attendanceCount int64
	s.db.Table("absensi").Where("sesi_id = ? AND warga_id = ?", sessionID, danruID).Count(&attendanceCount)
	if attendanceCount > 0 {
		return fmt.Errorf("Danru hadir, Wakil tidak dapat mengambil alih")
	}
	return nil
}

func validatePhoto(file multipart.File) (string, string, error) {
	buffer := make([]byte, 512)
	read, err := file.Read(buffer)
	if err != nil && read == 0 {
		return "", "", fmt.Errorf("file tidak dapat dibaca")
	}
	mimeType := http.DetectContentType(buffer[:read])
	extension, ok := map[string]string{"image/jpeg": ".jpg", "image/png": ".png"}[mimeType]
	if !ok {
		return "", "", fmt.Errorf("format foto harus JPG, JPEG, atau PNG")
	}
	return mimeType, extension, nil
}

func randomName(extension string) string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
	}
	return hex.EncodeToString(bytes) + extension
}

func hasRole(roles []string, target string) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}
