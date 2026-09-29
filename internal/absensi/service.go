package absensi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

type SessionView struct {
	ID           uint       `json:"id"`
	JadwalID     uint       `json:"jadwal_id"`
	ReguID       uint       `json:"regu_id"`
	TanggalRonda time.Time  `json:"tanggal_ronda"`
	MulaiAt      time.Time  `json:"mulai_at"`
	SelesaiAt    time.Time  `json:"selesai_at"`
	Status       string     `json:"status"`
	DimulaiPada  *time.Time `json:"dimulai_pada"`
	DitutupPada  *time.Time `json:"ditutup_pada"`
}

type AttendanceResult struct {
	AbsensiID uint   `json:"absensi_id"`
	Status    string `json:"status"`
	FotoID    uint   `json:"foto_id"`
}

type residentRef struct {
	ID     uint `gorm:"id"`
	ReguID uint `gorm:"regu_id"`
}

func (s *Service) List(userID uint, roles []string) ([]SessionView, error) {
	var daftar []SessionView
	query := s.db.Table("sesi_ronda AS sr").
		Select("sr.id, sr.jadwal_id, jr.regu_id, jr.tanggal_ronda, jr.mulai_at, jr.selesai_at, sr.status, sr.dimulai_pada, sr.diselesaikan_pada AS ditutup_pada").
		Joins("JOIN jadwal_ronda AS jr ON jr.id = sr.jadwal_id")
	if !hasRole(roles, "PAK_RT") && !hasRole(roles, "KOORDINATOR") {
		query = query.Joins("JOIN anggota_regu AS ar ON ar.regu_id = jr.regu_id").
			Joins("JOIN warga AS w ON w.id = ar.warga_id").
			Where("w.pengguna_id = ?", userID)
	}
	err := query.Order("jr.tanggal_ronda DESC, jr.mulai_at DESC").Scan(&daftar).Error
	return daftar, err
}

func (s *Service) Start(sessionID, userID uint, roles []string) error {
	if !hasRole(roles, "DANRU") && !hasRole(roles, "WAKIL_DANRU") {
		return fmt.Errorf("hanya Danru atau Wakil Danru yang dapat memulai sesi")
	}
	var count int64
	err := s.db.Table("petugas_regu AS pr").
		Joins("JOIN warga AS w ON w.id = pr.warga_id").
		Joins("JOIN sesi_ronda AS sr ON sr.id = ?", sessionID).
		Joins("JOIN jadwal_ronda AS jr ON jr.id = sr.jadwal_id AND jr.regu_id = pr.regu_id").
		Where("w.pengguna_id = ?", userID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("kamu bukan petugas sesi ini")
	}
	return s.db.Model(&models.SesiRonda{}).
		Where("id = ? AND status = ?", sessionID, "TERJADWAL").
		Updates(map[string]any{"status": "BERLANGSUNG", "dimulai_pada": time.Now()}).Error
}

func (s *Service) Create(sessionID, userID uint, roles []string, file multipart.File, header *multipart.FileHeader) (AttendanceResult, error) {
	if header.Size <= 0 || header.Size > s.cfg.MaxUploadBytes {
		return AttendanceResult{}, fmt.Errorf("ukuran file maksimal 5 MB")
	}
	mimeType, extension, err := validatePhoto(file)
	if err != nil {
		return AttendanceResult{}, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return AttendanceResult{}, fmt.Errorf("gagal membaca file")
	}

	var resident residentRef
	if err := s.db.Table("warga AS w").
		Select("w.id, ar.regu_id").
		Joins("JOIN anggota_regu AS ar ON ar.warga_id = w.id").
		Where("w.pengguna_id = ?", userID).
		Scan(&resident).Error; err != nil {
		return AttendanceResult{}, err
	}
	if resident.ID == 0 {
		return AttendanceResult{}, fmt.Errorf("warga belum memiliki regu")
	}

	var session models.SesiRonda
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return AttendanceResult{}, fmt.Errorf("sesi ronda tidak ditemukan")
	}
	var schedule models.JadwalRonda
	if err := s.db.First(&schedule, "id = ?", session.JadwalID).Error; err != nil {
		return AttendanceResult{}, err
	}
	if schedule.ReguID != resident.ReguID {
		return AttendanceResult{}, fmt.Errorf("sesi ini bukan untuk regu Anda")
	}
	if session.Status != "BERLANGSUNG" {
		return AttendanceResult{}, fmt.Errorf("absensi hanya dapat dilakukan saat sesi berlangsung")
	}
	now := time.Now()
	if now.Before(schedule.MulaiAt) || now.After(schedule.SelesaiAt) {
		return AttendanceResult{}, fmt.Errorf("waktu absensi di luar jadwal sesi")
	}

	isOfficer := false
	if hasRole(roles, "DANRU") || hasRole(roles, "WAKIL_DANRU") {
		var count int64
		if err := s.db.Table("petugas_regu AS pr").
			Joins("JOIN warga AS w ON w.id = pr.warga_id").
			Where("w.pengguna_id = ? AND pr.regu_id = ?", userID, schedule.ReguID).
			Count(&count).Error; err != nil {
			return AttendanceResult{}, err
		}
		isOfficer = count > 0
	}

	var existing models.Absensi
	result := s.db.Where("sesi_id = ? AND warga_id = ?", sessionID, resident.ID).Find(&existing)
	if result.Error != nil {
		return AttendanceResult{}, result.Error
	}
	if result.RowsAffected > 0 && existing.Status != "MENUNGGU" {
		return AttendanceResult{}, fmt.Errorf("absensi sudah final dan tidak dapat diubah")
	}

	photoPath, err := s.savePhoto(file, extension)
	if err != nil {
		return AttendanceResult{}, err
	}

	var foto models.FotoAbsensi
	if result.RowsAffected == 0 {
		existing = models.Absensi{SesiID: sessionID, WargaID: resident.ID, Status: "MENUNGGU", DiririmPada: now}
		if isOfficer {
			existing.Status = "TERVERIFIKASI"
		}
		if err := s.db.Create(&existing).Error; err != nil {
			_ = os.Remove(photoPath)
			return AttendanceResult{}, err
		}
		foto = models.FotoAbsensi{AbsensiID: existing.ID, LokasiFile: photoPath, NamaAsliFile: header.Filename, TipeMime: mimeType, UkuranFile: uint64(header.Size)}
		if err := s.db.Create(&foto).Error; err != nil {
			_ = os.Remove(photoPath)
			return AttendanceResult{}, err
		}
	} else {
		var old models.FotoAbsensi
		_ = s.db.First(&old, "absensi_id = ?", existing.ID).Error
		if err := s.db.Model(&existing).Update("dikirim_pada", now).Error; err != nil {
			_ = os.Remove(photoPath)
			return AttendanceResult{}, err
		}
		if err := s.db.Model(&models.FotoAbsensi{}).Where("absensi_id = ?", existing.ID).Updates(map[string]any{"lokasi_file": photoPath, "nama_asli_file": header.Filename, "tipe_mime": mimeType, "ukuran_file": header.Size}).Error; err != nil {
			_ = os.Remove(photoPath)
			return AttendanceResult{}, err
		}
		if old.LokasiFile != "" {
			_ = os.Remove(old.LokasiFile)
		}
		_ = s.db.First(&foto, "absensi_id = ?", existing.ID).Error
	}
	return AttendanceResult{AbsensiID: existing.ID, Status: existing.Status, FotoID: foto.ID}, nil
}

func (s *Service) savePhoto(file multipart.File, extension string) (string, error) {
	dir := filepath.Join(s.cfg.UploadDir, "absensi")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("gagal menyiapkan folder evidence")
	}
	path := filepath.Join(dir, randomName(extension))
	destination, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan evidence")
	}
	defer destination.Close()
	if _, err := io.Copy(destination, file); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("gagal menyimpan evidence")
	}
	return path, nil
}

func validatePhoto(file multipart.File) (string, string, error) {
	buffer := make([]byte, 512)
	read, err := file.Read(buffer)
	if err != nil && read == 0 {
		return "", "", fmt.Errorf("file foto tidak dapat dibaca")
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
