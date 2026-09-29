package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"sispat/backend/internal/config"
	"sispat/backend/internal/models"
)

type UserAccess struct {
	PenggunaID uint     `json:"pengguna_id"`
	Username   string   `json:"nama_pengguna"`
	Roles      []string `json:"peran"`
	CSRFToken  string   `json:"csrf_token"`
}

type Service struct {
	db            *gorm.DB
	config        config.Config
	limitMu       sync.Mutex
	loginAttempts map[string]loginAttempt
}

type loginAttempt struct {
	Count   int
	ResetAt time.Time
}

func NewService(db *gorm.DB, cfg config.Config) *Service {
	return &Service{db: db, config: cfg, loginAttempts: make(map[string]loginAttempt)}
}

func (s *Service) Login(username, password, ipAddress, userAgent string) (UserAccess, string, error) {
	username = normalizeUsername(username)
	attemptKey := ipAddress + ":" + username
	if !s.allowLoginAttempt(attemptKey) {
		return UserAccess{}, "", errors.New("terlalu banyak percobaan login")
	}
	pengguna, err := s.findPengguna(username)
	if err != nil {
		log.Printf("login gagal: pengguna %q tidak ditemukan", username)
		s.registerLoginFailure(attemptKey)
		return UserAccess{}, "", err
	}

	if !pengguna.Aktif {
		log.Printf("login gagal: pengguna %q tidak aktif", username)
		s.registerLoginFailure(attemptKey)
		return UserAccess{}, "", errors.New("kredensial tidak valid")
	}

	if !VerifyPassword(pengguna.KataSandiHash, password) {
		log.Printf("login gagal: kata sandi salah untuk pengguna %q", username)
		s.registerLoginFailure(attemptKey)
		return UserAccess{}, "", errors.New("kredensial tidak valid")
	}

	access, err := s.userAccess(pengguna.ID, pengguna.NamaPengguna)
	if err != nil {
		return UserAccess{}, "", err
	}

	sessionID, err := newRandomToken()
	if err != nil {
		return UserAccess{}, "", err
	}
	csrfToken, err := newRandomToken()
	if err != nil {
		return UserAccess{}, "", err
	}

	now := time.Now()
	expiredAt := now.Add(time.Duration(s.config.SessionTTLHours) * time.Hour)
	session := models.SesiPengguna{
		ID:                sessionID,
		CSRFToken:         csrfToken,
		PenggunaID:        pengguna.ID,
		DibuatPada:        now,
		TerakhirAktifPada: now,
		KedaluwarsaPada:   expiredAt,
		AlamatIP:          truncate(ipAddress, 45),
		UserAgent:         truncate(userAgent, 500),
	}
	if err := s.db.Create(&session).Error; err != nil {
		return UserAccess{}, "", err
	}

	nowTime := now
	if err := s.db.Model(&pengguna).Update("terakhir_masuk_pada", &nowTime).Error; err != nil {
		return UserAccess{}, "", err
	}

	access.CSRFToken = csrfToken
	s.resetLoginAttempt(attemptKey)
	return access, sessionID, nil
}

func (s *Service) Logout(sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.db.Delete(&models.SesiPengguna{}, "id = ?", sessionID).Error
}

func (s *Service) CurrentUser(sessionID string) (UserAccess, error) {
	if sessionID == "" {
		return UserAccess{}, sql.ErrNoRows
	}

	var session models.SesiPengguna
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return UserAccess{}, err
	}
	if time.Now().After(session.KedaluwarsaPada) {
		_ = s.db.Delete(&models.SesiPengguna{}, "id = ?", sessionID).Error
		return UserAccess{}, sql.ErrNoRows
	}

	now := time.Now()
	if err := s.db.Model(&session).Update("terakhir_aktif_pada", now).Error; err != nil {
		return UserAccess{}, err
	}

	var pengguna models.Pengguna
	if err := s.db.First(&pengguna, session.PenggunaID).Error; err != nil {
		return UserAccess{}, err
	}
	if !pengguna.Aktif {
		return UserAccess{}, sql.ErrNoRows
	}

	access, err := s.userAccess(pengguna.ID, pengguna.NamaPengguna)
	if err != nil {
		return UserAccess{}, err
	}
	access.CSRFToken = session.CSRFToken
	return access, nil
}

func (s *Service) userAccess(penggunaID uint, username string) (UserAccess, error) {
	access := UserAccess{
		PenggunaID: penggunaID,
		Username:   username,
		Roles:      make([]string, 0, 3),
	}

	var roles []string
	if err := s.db.Raw(`
		SELECT p.kode
		FROM pengguna_peran pp
		JOIN peran p ON p.id = pp.peran_id
		WHERE pp.pengguna_id = ?
	`, penggunaID).Scan(&roles).Error; err != nil {
		return UserAccess{}, err
	}
	access.Roles = append(access.Roles, roles...)

	var office string
	if err := s.db.Raw(`
		SELECT pr.jabatan
		FROM petugas_regu pr
		JOIN warga w ON w.id = pr.warga_id
		WHERE w.pengguna_id = ?
		LIMIT 1
	`, penggunaID).Scan(&office).Error; err != nil {
		return UserAccess{}, err
	}
	if office != "" {
		access.Roles = append(access.Roles, office)
	}

	return access, nil
}

func (s *Service) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.config.SessionCookie,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.config.SessionSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   s.config.SessionTTLHours * 3600,
	})
}

func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.config.SessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.config.SessionSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (s *Service) SessionID(r *http.Request) string {
	cookie, err := r.Cookie(s.config.SessionCookie)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

type contextKey string

const userContextKey contextKey = "sispat-user"

func WithUser(ctx context.Context, user UserAccess) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (UserAccess, bool) {
	user, ok := ctx.Value(userContextKey).(UserAccess)
	return user, ok
}

// allowLoginAttempt hanya mengecek apakah percobaan masih diizinkan.
// Penghitungan kegagalan dilakukan pada registerLoginFailure agar percobaan
// berhasil tidak ikut membatasi login berikutnya.
func (s *Service) allowLoginAttempt(key string) bool {
	s.limitMu.Lock()
	defer s.limitMu.Unlock()
	now := time.Now()
	attempt := s.loginAttempts[key]
	if attempt.ResetAt.Before(now) {
		return true
	}
	return attempt.Count < 5
}

func (s *Service) registerLoginFailure(key string) {
	s.limitMu.Lock()
	defer s.limitMu.Unlock()
	now := time.Now()
	attempt := s.loginAttempts[key]
	if attempt.ResetAt.Before(now) {
		attempt = loginAttempt{ResetAt: now.Add(time.Minute)}
	}
	attempt.Count++
	s.loginAttempts[key] = attempt
}

func (s *Service) resetLoginAttempt(key string) {
	s.limitMu.Lock()
	defer s.limitMu.Unlock()
	delete(s.loginAttempts, key)
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// findPengguna mencari pengguna berdasarkan nama_pengguna yang sudah dinormalisasi.
// Pencarian memakai LOWER(...) supaya tidak bergantung pada collation MySQL.
func (s *Service) findPengguna(username string) (models.Pengguna, error) {
	var pengguna models.Pengguna
	err := s.db.Where("LOWER(nama_pengguna) = ?", username).First(&pengguna).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Pengguna{}, errors.New("kredensial tidak valid")
	}
	if err != nil {
		return models.Pengguna{}, err
	}
	return pengguna, nil
}

func newRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
