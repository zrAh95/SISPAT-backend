package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"sispat/backend/internal/absensi"
	"sispat/backend/internal/auth"
	"sispat/backend/internal/config"
	"sispat/backend/internal/dashboard"
	"sispat/backend/internal/database"
	httpapi "sispat/backend/internal/http"
	"sispat/backend/internal/jadwal"
	"sispat/backend/internal/koordinator"
	"sispat/backend/internal/laporan"
	"sispat/backend/internal/regu"
	"sispat/backend/internal/verifikasi"
	"sispat/backend/internal/warga"
)

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
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("gagal mengecek database: %v", err)
	}
	defer sqlDB.Close()

	authService := auth.NewService(db, cfg)
	handler := httpapi.NewHandler(authService)
	wargaService := warga.NewService(db)
	wargaHandler := httpapi.NewWargaHandler(wargaService)
	reguService := regu.NewService(db)
	reguHandler := httpapi.NewReguHandler(reguService)
	koordinatorService := koordinator.NewService(db)
	koordinatorHandler := httpapi.NewKoordinatorHandler(koordinatorService)
	jadwalService := jadwal.NewService(db)
	jadwalHandler := httpapi.NewJadwalHandler(jadwalService)
	absensiService := absensi.NewService(db, cfg)
	absensiHandler := httpapi.NewAbsensiHandler(absensiService, cfg)
	verifikasiService := verifikasi.NewService(db)
	verifikasiHandler := httpapi.NewVerifikasiHandler(verifikasiService)
	laporanService := laporan.NewService(db, cfg)
	laporanHandler := httpapi.NewLaporanHandler(laporanService, cfg)
	dashboardService := dashboard.NewService(db)
	dashboardHandler := httpapi.NewDashboardHandler(dashboardService)

	router := chi.NewRouter()
	router.Use(httpapi.Recovery)
	router.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"layanan": "SISPAT API",
		})
	})

	router.Post("/api/auth/login", handler.Login)
	router.Group(func(r chi.Router) {
		r.Use(authService.RequireSession)
		r.Post("/api/auth/logout", handler.Logout)
		r.Get("/api/auth/me", handler.Me)
		r.Get("/api/sesi-ronda", absensiHandler.ListSessions)
		r.Post("/api/sesi-ronda/{id}/absensi", absensiHandler.Create)
		r.Get("/api/foto-absensi/{id}", verifikasiHandler.Photo)

		r.Group(func(r chi.Router) {
			r.Use(authService.RequireRoles("DANRU", "WAKIL_DANRU"))
			r.Post("/api/sesi-ronda/{id}/mulai", absensiHandler.Start)
			r.Get("/api/sesi-ronda/{id}/absensi", verifikasiHandler.List)
			r.Post("/api/absensi/{id}/verifikasi", verifikasiHandler.Verify)
			r.Post("/api/absensi/{id}/tolak", verifikasiHandler.Reject)
			r.Post("/api/sesi-ronda/{id}/laporan", laporanHandler.Create)
			r.Post("/api/laporan/{id}/bukti", laporanHandler.Evidence)
		})

		r.Get("/api/dashboard", dashboardHandler.Summary)

		r.Group(func(r chi.Router) {
			r.Use(authService.RequireRoles("PAK_RT", "KOORDINATOR"))
			r.Get("/api/dashboard/kehadiran", dashboardHandler.Kehadiran)
			r.Get("/api/rekap", dashboardHandler.Rekap)
		})

		// Pembacaan regu dibuka untuk Pak RT dan Koordinator.
		r.Group(func(r chi.Router) {
			r.Use(authService.RequireRoles("PAK_RT", "KOORDINATOR"))
			r.Get("/api/regu", reguHandler.List)
			r.Get("/api/regu/{id}/anggota", reguHandler.Members)
			r.Get("/api/regu/{id}/petugas", reguHandler.Officers)
		})

		r.Group(func(r chi.Router) {
			r.Use(authService.RequireRoles("PAK_RT"))
			r.Get("/api/jadwal-ronda", jadwalHandler.List)
			r.Post("/api/jadwal-ronda", jadwalHandler.Create)
			r.Get("/api/jadwal-ronda/{id}", jadwalHandler.Get)
			r.Get("/api/koordinator/calon", koordinatorHandler.ListCalon)
			r.Post("/api/koordinator", koordinatorHandler.Assign)
		})

		r.Group(func(r chi.Router) {
			r.Use(authService.RequireRoles("KOORDINATOR"))
			r.Get("/api/warga", wargaHandler.List)
			r.Post("/api/warga", wargaHandler.Create)
			r.Get("/api/warga/{id}", wargaHandler.Get)
			r.Put("/api/warga/{id}", wargaHandler.Update)
			r.Delete("/api/warga/{id}", wargaHandler.Delete)
			r.Patch("/api/warga/{id}/status", wargaHandler.SetStatus)

			r.Post("/api/regu/{id}/danru", reguHandler.AssignDanru)
			r.Post("/api/regu/{id}/wakil-danru", reguHandler.AssignWakil)
		})
	})

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("SISPAT API berjalan di http://localhost:%s", cfg.AppPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
