package verifikasi

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"

	"sispat/backend/internal/models"
)

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

type Item struct {
	AbsensiID   uint      `json:"absensi_id"`
	WargaID     uint      `json:"warga_id"`
	Nama        string    `json:"nama"`
	Status      string    `json:"status"`
	DiririmPada time.Time `json:"dikirim_pada"`
	FotoID      *uint     `json:"foto_id"`
	FotoAda     bool      `json:"foto_ada"`
}

func (s *Service) ListSession(sessionID, userID uint, roles []string) ([]Item, error) {
	if !hasRole(roles, "DANRU") && !hasRole(roles, "WAKIL_DANRU") {
		return nil, fmt.Errorf("hanya Danru atau Wakil Danru yang dapat melihat verifikasi")
	}
	groupID, err := s.sessionGroup(sessionID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeOfficer(userID, groupID, roles, sessionID); err != nil {
		return nil, err
	}

	var items []Item
	err = s.db.Table("absensi AS a").
		Select("a.id AS absensi_id, a.warga_id, w.nama, a.status, a.dikirim_pada, f.id AS foto_id, CASE WHEN f.id IS NULL THEN FALSE ELSE TRUE END AS foto_ada").
		Joins("JOIN warga AS w ON w.id = a.warga_id").
		Joins("LEFT JOIN foto_absensi AS f ON f.absensi_id = a.id").
		Joins("JOIN sesi_ronda AS s ON s.id = a.sesi_id").
		Joins("JOIN jadwal_ronda AS j ON j.id = s.jadwal_id AND j.regu_id = ?", groupID).
		Where("a.sesi_id = ?", sessionID).
		Order("w.nama ASC").
		Scan(&items).Error
	return items, err
}

func (s *Service) Verify(attendanceID, userID uint, roles []string, reject bool, reason string) error {
	if reject && strings.TrimSpace(reason) == "" {
		return fmt.Errorf("alasan penolakan wajib diisi")
	}
	if !hasRole(roles, "DANRU") && !hasRole(roles, "WAKIL_DANRU") {
		return fmt.Errorf("hanya Danru atau Wakil Danru yang dapat memverifikasi")
	}

	var attendance models.Absensi
	if err := s.db.First(&attendance, "id = ?", attendanceID).Error; err != nil {
		return fmt.Errorf("absensi tidak ditemukan")
	}
	if attendance.Status != "MENUNGGU" {
		return fmt.Errorf("hanya absensi MENUNGGU yang dapat diproses")
	}
	groupID, err := s.sessionGroupForAttendance(attendance)
	if err != nil {
		return err
	}
	if err := s.authorizeOfficer(userID, groupID, roles, attendance.SesiID); err != nil {
		return err
	}

	status := "TERVERIFIKASI"
	if reject {
		status = "DITOLAK"
	}
	now := time.Now()
	return s.db.Model(&attendance).Updates(map[string]any{
		"status":                        status,
		"diverifikasi_oleh_pengguna_id": userID,
		"diverifikasi_pada":             now,
		"alasan_ditolak":                strings.TrimSpace(reason),
	}).Error
}

func (s *Service) OpenPhoto(photoID, userID uint, roles []string) (string, string, error) {
	var photo models.FotoAbsensi
	if err := s.db.First(&photo, "id = ?", photoID).Error; err != nil {
		return "", "", fmt.Errorf("foto tidak ditemukan")
	}
	var attendance models.Absensi
	if err := s.db.First(&attendance, "id = ?", photo.AbsensiID).Error; err != nil {
		return "", "", fmt.Errorf("absensi tidak ditemukan")
	}
	groupID, err := s.sessionGroupForAttendance(attendance)
	if err != nil {
		return "", "", err
	}
	allowed := hasRole(roles, "PAK_RT") || hasRole(roles, "KOORDINATOR")
	if hasRole(roles, "DANRU") || hasRole(roles, "WAKIL_DANRU") {
		allowed = true
	}
	if !allowed {
		var own int64
		s.db.Table("warga AS w").Joins("JOIN anggota_regu AS ar ON ar.warga_id = w.id").Where("w.pengguna_id = ? AND ar.regu_id = ?", userID, groupID).Count(&own)
		allowed = own > 0
	}
	if !allowed {
		return "", "", fmt.Errorf("Anda tidak memiliki akses ke foto ini")
	}
	path := photo.LokasiFile
	if _, err := os.Stat(path); err != nil {
		return "", "", fmt.Errorf("file foto tidak ditemukan")
	}
	return path, photo.TipeMime, nil
}

func (s *Service) sessionGroup(sessionID uint) (uint, error) {
	var groupID uint
	err := s.db.Table("sesi_ronda AS s").Select("j.regu_id").Joins("JOIN jadwal_ronda AS j ON j.id = s.jadwal_id").Where("s.id = ?", sessionID).Scan(&groupID).Error
	if err != nil || groupID == 0 {
		return 0, fmt.Errorf("sesi ronda tidak ditemukan")
	}
	return groupID, nil
}

func (s *Service) sessionGroupForAttendance(a models.Absensi) (uint, error) {
	return s.sessionGroup(a.SesiID)
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
	if err := s.db.Table("petugas_regu AS pr").Joins("JOIN warga AS w ON w.id = pr.warga_id").Where("pr.regu_id = ? AND pr.jabatan = ?", groupID, "DANRU").Select("w.id").Scan(&danruID).Error; err != nil {
		return err
	}
	if danruID == 0 {
		return fmt.Errorf("Danru belum ditetapkan")
	}
	var countAbsensi int64
	s.db.Table("absensi").Where("sesi_id = ? AND warga_id = ?", sessionID, danruID).Count(&countAbsensi)
	if countAbsensi > 0 {
		return fmt.Errorf("Danru hadir, Wakil tidak dapat mengambil alih verifikasi")
	}
	return nil
}

func hasRole(roles []string, target string) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}
