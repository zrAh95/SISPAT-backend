package dashboard

import (
	"time"

	"gorm.io/gorm"

	"sispat/backend/internal/models"
)

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

type Summary struct {
	JumlahWarga       int64   `json:"jumlah_warga"`
	JumlahRegu        int64   `json:"jumlah_regu"`
	JumlahRonda       int64   `json:"jumlah_ronda"`
	RondaSelesai      int64   `json:"ronda_selesai"`
	RondaBelumSelesai int64   `json:"ronda_belum_selesai"`
	Hadir             int64   `json:"hadir"`
	BelumHadir        int64   `json:"belum_hadir"`
	Pending           int64   `json:"pending"`
	TotalJimpitan     float64 `json:"total_jimpitan"`
}

type ReguAttendance struct {
	Kode          string `json:"kode"`
	Nama          string `json:"nama"`
	JumlahAnggota int64  `json:"jumlah_anggota"`
	Hadir         int64  `json:"hadir"`
	BelumHadir    int64  `json:"belum_hadir"`
	Pending       int64  `json:"pending"`
}

type Rekap struct {
	ID                uint      `json:"id"`
	JadwalID          uint      `json:"jadwal_id"`
	ReguID            uint      `json:"regu_id"`
	Regu              string    `json:"regu"`
	TanggalRonda      time.Time `json:"tanggal_ronda"`
	MulaiAt           time.Time `json:"mulai_at"`
	SelesaiAt         time.Time `json:"selesai_at"`
	StatusSesi        string    `json:"status_sesi"`
	JumlahAnggota     int64     `json:"jumlah_anggota"`
	Hadir             int64     `json:"hadir"`
	BelumHadir        int64     `json:"belum_hadir"`
	Pending           int64     `json:"pending"`
	StatusLaporan     string    `json:"status_laporan"`
	RingkasanJimpitan string    `json:"ringkasan_jimpitan"`
}

func (s *Service) Summary(userID uint, roles []string) (Summary, error) {
	var result Summary
	groupID := s.userGroup(userID, roles)
	warga := s.db.Model(&models.Warga{}).Where("aktif = TRUE")
	sesi := s.db.Table("sesi_ronda AS sr").Joins("JOIN jadwal_ronda AS jr ON jr.id = sr.jadwal_id")
	absensi := s.db.Table("absensi AS a").Joins("JOIN sesi_ronda AS sr ON sr.id = a.sesi_id").Joins("JOIN jadwal_ronda AS jr ON jr.id = sr.jadwal_id")
	laporan := s.db.Table("laporan_ronda AS lr").Joins("JOIN sesi_ronda AS sr ON sr.id = lr.sesi_id").Joins("JOIN jadwal_ronda AS jr ON jr.id = sr.jadwal_id")
	if groupID != 0 {
		warga = warga.Where("id IN (SELECT warga_id FROM anggota_regu WHERE regu_id = ?)", groupID)
		sesi = sesi.Where("jr.regu_id = ?", groupID)
		absensi = absensi.Where("jr.regu_id = ?", groupID)
		laporan = laporan.Where("jr.regu_id = ?", groupID)
	}
	if err := warga.Count(&result.JumlahWarga).Error; err != nil {
		return Summary{}, err
	}
	s.db.Model(&models.Regu{}).Count(&result.JumlahRegu)
	sesi.Count(&result.JumlahRonda)
	countSessions := s.db.Table("sesi_ronda AS sr").Joins("JOIN jadwal_ronda AS jr ON jr.id = sr.jadwal_id")
	if groupID != 0 {
		countSessions = countSessions.Where("jr.regu_id = ?", groupID)
	}
	countSessions.Where("sr.status = 'SELESAI'").Count(&result.RondaSelesai)
	countSessions = s.db.Table("sesi_ronda AS sr").Joins("JOIN jadwal_ronda AS jr ON jr.id = sr.jadwal_id")
	if groupID != 0 {
		countSessions = countSessions.Where("jr.regu_id = ?", groupID)
	}
	countSessions.Where("sr.status <> 'SELESAI'").Count(&result.RondaBelumSelesai)
	absensi.Where("a.status = 'TERVERIFIKASI'").Count(&result.Hadir)
	absensi.Where("a.status = 'MENUNGGU'").Count(&result.Pending)
	result.BelumHadir = result.JumlahWarga - result.Hadir - result.Pending
	laporan.Select("COALESCE(SUM(lr.total_jimpitan), 0)").Scan(&result.TotalJimpitan)
	return result, nil
}

func (s *Service) KehadiranPerRegu() ([]ReguAttendance, error) {
	var result []ReguAttendance
	err := s.db.Raw(`
		SELECT r.kode, r.nama,
			COUNT(DISTINCT ar.warga_id) AS jumlah_anggota,
			COUNT(DISTINCT CASE WHEN a.status = 'TERVERIFIKASI' THEN a.warga_id END) AS hadir,
			COUNT(DISTINCT CASE WHEN a.status = 'MENUNGGU' THEN a.warga_id END) AS pending,
			COUNT(DISTINCT ar.warga_id) - COUNT(DISTINCT CASE WHEN a.status IN ('TERVERIFIKASI', 'MENUNGGU') THEN a.warga_id END) AS belum_hadir
		FROM regu r
		LEFT JOIN anggota_regu ar ON ar.regu_id = r.id
		LEFT JOIN (
			SELECT jr.regu_id, MAX(sr.id) AS sesi_id
			FROM sesi_ronda sr
			JOIN jadwal_ronda jr ON jr.id = sr.jadwal_id
			GROUP BY jr.regu_id
		) latest ON latest.regu_id = r.id
		LEFT JOIN absensi a ON a.sesi_id = latest.sesi_id
		GROUP BY r.id, r.kode, r.nama
		ORDER BY r.kode ASC
	`).Scan(&result).Error
	if result == nil {
		result = make([]ReguAttendance, 0)
	}
	return result, err
}

func (s *Service) Rekap(userID uint, roles []string, start, end string, reguID uint) ([]Rekap, error) {
	query := s.db.Table("sesi_ronda AS sr").
		Select("sr.id, sr.jadwal_id, jr.regu_id, r.nama AS regu, jr.tanggal_ronda, jr.mulai_at, jr.selesai_at, sr.status AS status_sesi, COUNT(DISTINCT ar.warga_id) AS jumlah_anggota, SUM(CASE WHEN a.status = 'TERVERIFIKASI' THEN 1 ELSE 0 END) AS hadir, SUM(CASE WHEN a.status = 'MENUNGGU' THEN 1 ELSE 0 END) AS pending, COALESCE(lr.status_kegiatan, '') AS status_laporan, COALESCE(lr.ringkasan_jimpitan, '') AS ringkasan_jimpitan").
		Joins("JOIN jadwal_ronda AS jr ON jr.id = sr.jadwal_id").
		Joins("JOIN regu AS r ON r.id = jr.regu_id").
		Joins("LEFT JOIN anggota_regu AS ar ON ar.regu_id = jr.regu_id").
		Joins("LEFT JOIN absensi AS a ON a.sesi_id = sr.id AND a.warga_id = ar.warga_id").
		Joins("LEFT JOIN laporan_ronda AS lr ON lr.sesi_id = sr.id")
	if start != "" {
		query = query.Where("jr.tanggal_ronda >= ?", start)
	}
	if end != "" {
		query = query.Where("jr.tanggal_ronda <= ?", end)
	}
	if reguID != 0 {
		query = query.Where("jr.regu_id = ?", reguID)
	} else if !hasRole(roles, "PAK_RT") && !hasRole(roles, "KOORDINATOR") {
		query = query.Where("jr.regu_id = ?", s.userGroup(userID, roles))
	}
	result := make([]Rekap, 0)
	err := query.Group("sr.id, sr.jadwal_id, jr.regu_id, r.nama, jr.tanggal_ronda, jr.mulai_at, jr.selesai_at, sr.status, lr.status_kegiatan, lr.ringkasan_jimpitan").Order("jr.tanggal_ronda DESC").Scan(&result).Error
	return result, err
}

func (s *Service) userGroup(userID uint, roles []string) uint {
	if hasRole(roles, "PAK_RT") || hasRole(roles, "KOORDINATOR") {
		return 0
	}
	var id uint
	s.db.Table("anggota_regu AS ar").Joins("JOIN warga AS w ON w.id = ar.warga_id").Where("w.pengguna_id = ?", userID).Select("ar.regu_id").Scan(&id)
	return id
}
func hasRole(roles []string, target string) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}
