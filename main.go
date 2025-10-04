package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	"smartseller-lite-starter/internal/app"
	"smartseller-lite-starter/internal/db"
	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/httpapi"
	"smartseller-lite-starter/internal/media"
	"smartseller-lite-starter/internal/service"
	"smartseller-lite-starter/internal/util"
)

//go:embed frontend/dist/*
var embeddedDist embed.FS

const (
	healthEndpoint      = "/api/health"
	healthCheckAttempts = 20
	healthCheckInterval = 200 * time.Millisecond
)

func loadEnvFiles() {
	envPaths := []string{".env"}

	if runtime.GOOS == "windows" {
		programData := strings.TrimSpace(os.Getenv("PROGRAMDATA"))
		if programData != "" {
			envPaths = append(envPaths, filepath.Join(programData, "SmartSellerLite", "config", ".env"))
		}
	}

	for _, envPath := range envPaths {
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Overload(envPath); err != nil {
				log.Printf("load env file %s: %v", envPath, err)
			}
		}
	}
}

func main() {
	loadEnvFiles()

	var (
		openBrowserFlag bool
		addr            string
		dsn             string
	)

	defaultOpenBrowser := runtime.GOOS == "windows"
	flag.BoolVar(&openBrowserFlag, "open-browser", defaultOpenBrowser, "open default browser after server starts")
	flag.StringVar(&addr, "addr", getEnv("APP_ADDR", "127.0.0.1:8787"), "HTTP bind address")
	defaultDSN := strings.TrimSpace(getEnv("DATABASE_DSN", ""))
	if defaultDSN == "" {
		defaultDSN = "root:smartseller@tcp(127.0.0.1:3306)/smartseller"
	}
	flag.StringVar(&dsn, "dsn", defaultDSN, "MySQL DSN (e.g. user:pass@tcp(127.0.0.1:3306)/smartseller)")
	flag.Parse()

	openURL := fmt.Sprintf("http://%s/", addr)
	healthURL := fmt.Sprintf("http://%s%s", addr, healthEndpoint)

	if !portAvailable(addr) {
		if openBrowserFlag {
			client := &http.Client{Timeout: time.Second}
			if resp, err := client.Get(healthURL); err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					util.OpenBrowser(openURL)
					return
				}
			}
			util.OpenBrowser(openURL)
		}
		log.Printf("address %s already in use, exiting", addr)
		return
	}

	log.Printf("mariadb dsn: %v", dsn)
	log.Printf("default dsn: %v", defaultDSN)

	if err := ensureLocalMariaDB(dsn); err != nil {
		log.Printf("mariadb bootstrap: %v", err)
		util.ShowFatalError("SmartSeller Lite", fmt.Sprintf("Gagal menyiapkan layanan MariaDB lokal. Jalankan aplikasi sebagai Administrator atau instal ulang dengan opsi bundled DB.\n\nDetail error:\n%v", err))
	}

	store, err := db.NewStore(strings.TrimSpace(dsn))
	if err != nil {
		log.Printf("init database: %v", err)

		var authErr *db.UnsupportedAuthPluginError
		if errors.As(err, &authErr) {
			message := buildAuthPluginErrorMessage(dsn, authErr.Plugin)
			util.ShowFatalError("SmartSeller Lite", message)
			return
		}

		errMessage := "SmartSeller Lite tidak bisa tersambung ke database. Pastikan layanan MariaDB berjalan atau periksa ulang credential yang disimpan saat proses instalasi."
		util.ShowFatalError("SmartSeller Lite", fmt.Sprintf("%s\n\nDetail error:\n%v", errMessage, err))
		return
	}
	brandName := getEnv("APP_BRAND_NAME", "SmartSeller Lite")

	mediaBase, err := media.ResolveBaseDir()
	if err != nil {
		log.Fatalf("resolve media base: %v", err)
	}
	mediaManager, err := media.NewManager(mediaBase)
	if err != nil {
		log.Fatalf("prepare media manager: %v", err)
	}

	core := app.NewCore(store, app.CoreConfig{DefaultBrandName: brandName, MediaManager: mediaManager})
	core.Warm(context.Background())
	defer func() {
		if err := core.Close(); err != nil {
			log.Printf("close store: %v", err)
		}
	}()

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	api := app.NewAPI(core)
	mountAPI(router, api)
	router.Get(healthEndpoint, httpapi.HealthHandler(brandName, store))

	router.Handle("/media/*", http.StripPrefix("/media/", http.FileServer(http.Dir(mediaManager.MediaDir()))))

	dist, source, err := resolveFrontendDist()
	if err != nil {
		log.Printf("frontend assets unavailable: %v", err)
		router.Handle("/*", newMissingFrontendHandler(err))
	} else {
		log.Printf("serving frontend assets from %s", source)
		router.Handle("/*", newSPAHandler(dist))
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("SmartSeller Lite listening on %s (db: %s, media: %s)\n", openURL, store.RedactedDSN(), mediaManager.MediaDir())

	if openBrowserFlag {
		go waitForHealthAndOpen(healthURL, openURL)
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func parsePositiveInt(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func waitForHealthAndOpen(healthURL, openURL string) {
	client := &http.Client{Timeout: time.Second}
	for attempt := 0; attempt < healthCheckAttempts; attempt++ {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				util.OpenBrowser(openURL)
				return
			}
		}
		time.Sleep(healthCheckInterval)
	}
}
func buildAuthPluginErrorMessage(rawDSN, plugin string) string {
	pluginName := strings.TrimSpace(plugin)
	if pluginName == "" {
		pluginName = "non-standar"
	}

	cfg, err := mysqlDriver.ParseDSN(strings.TrimSpace(rawDSN))
	user := "<user>"
	host := "localhost"
	if err == nil {
		if u := strings.TrimSpace(cfg.User); u != "" {
			user = u
		}
		if addr := strings.TrimSpace(cfg.Addr); addr != "" {
			host = addr
			if idx := strings.Index(host, ":"); idx >= 0 {
				host = host[:idx]
			}
			host = strings.TrimSpace(host)
			if host == "" {
				host = "localhost"
			}
		}
	}

	var b strings.Builder
	b.WriteString("SmartSeller Lite gagal login ke MariaDB karena server meminta plugin otentikasi \"")
	b.WriteString(pluginName)
	b.WriteString("\" yang belum didukung oleh driver MySQL bawaan aplikasi.\n\n")

	b.WriteString("Langkah yang disarankan:\n")
	b.WriteString("  1. Buka command prompt, masuk sebagai admin ke MariaDB (contoh: `mysql -u root -p`).\n")
	b.WriteString("  2. Jalankan perintah berikut untuk akun yang dipakai aplikasi:\n\n")
	b.WriteString("     ALTER USER '")
	b.WriteString(user)
	b.WriteString("'@'")
	b.WriteString(host)
	b.WriteString("' IDENTIFIED WITH mysql_native_password BY '<password-saat-ini>';\n\n")
	b.WriteString("     Ganti <password-saat-ini> dengan kata sandi yang tersimpan di DSN aplikasi (file .env).\n")
	b.WriteString("  3. Restart layanan MariaDB lalu buka kembali SmartSeller Lite.\n\n")
	b.WriteString("Alternatif: set parameter `default_authentication_plugin=mysql_native_password` pada konfigurasi MariaDB agar akun baru otomatis memakai plugin tersebut.\n")

	return b.String()
}

func ensureLocalMariaDB(dsn string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	if !shouldManageLocalMariaDB(dsn) {
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	mysqldPath := filepath.Join(exeDir, "mariadb", "bin", "mysqld.exe")
	if !fileExists(mysqldPath) {
		return nil
	}

	programData := strings.TrimSpace(os.Getenv("PROGRAMDATA"))
	if programData == "" {
		return errors.New("CSIDL_COMMON_APPDATA environment variable is not set")
	}

	mysqlDir := filepath.Join(programData, "SmartSellerLite", "mysql")
	dataDir := filepath.Join(programData, "SmartSellerLite", "data")
	if err := os.MkdirAll(mysqlDir, 0o755); err != nil {
		return fmt.Errorf("create mysql config dir: %w", err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create mysql data dir: %w", err)
	}

	defaultsFile := filepath.Join(mysqlDir, "my.ini")
	if !fileExists(defaultsFile) {
		return fmt.Errorf("file my.ini tidak ditemukan di %s", defaultsFile)
	}

	if err := ensureWindowsServiceRunning("SmartSellerDB", mysqldPath, defaultsFile); err != nil {
		return err
	}

	return nil
}

func ensureWindowsServiceRunning(name, mysqldPath, defaultsFile string) error {
	var buf bytes.Buffer

	query := exec.Command("sc.exe", "query", name)
	query.Stdout = &buf
	query.Stderr = &buf
	err := query.Run()
	output := buf.String()
	if err == nil {
		if strings.Contains(strings.ToUpper(output), "RUNNING") {
			return nil
		}

		buf.Reset()
		start := exec.Command("sc.exe", "start", name)
		start.Stdout = &buf
		start.Stderr = &buf
		if err := start.Run(); err != nil {
			startOutput := buf.String()
			if !strings.Contains(startOutput, "1056") {
				return fmt.Errorf("start service %s: %s", name, strings.TrimSpace(startOutput))
			}
		}
		return nil
	}

	if !strings.Contains(output, "1060") && !strings.Contains(err.Error(), "1060") {
		return fmt.Errorf("query service %s: %s", name, strings.TrimSpace(output))
	}

	buf.Reset()
	install := exec.Command(mysqldPath, "--install", name, "--defaults-file="+defaultsFile)
	install.Stdout = &buf
	install.Stderr = &buf
	if err := install.Run(); err != nil {
		return fmt.Errorf("install service %s: %s", name, strings.TrimSpace(buf.String()))
	}

	buf.Reset()
	start := exec.Command("sc.exe", "start", name)
	start.Stdout = &buf
	start.Stderr = &buf
	if err := start.Run(); err != nil {
		startOutput := buf.String()
		if !strings.Contains(startOutput, "1056") {
			return fmt.Errorf("start service %s: %s", name, strings.TrimSpace(startOutput))
		}
	}

	return nil
}

func shouldManageLocalMariaDB(dsn string) bool {
	raw := strings.ToLower(strings.TrimSpace(dsn))
	if raw == "" {
		return true
	}

	idx := strings.Index(raw, "@tcp(")
	if idx == -1 {
		return false
	}

	hostPort := raw[idx+5:]
	closing := strings.Index(hostPort, ")")
	if closing == -1 {
		return false
	}

	hostPort = hostPort[:closing]
	host := hostPort
	if colon := strings.Index(hostPort, ":"); colon >= 0 {
		host = hostPort[:colon]
	}
	host = strings.TrimSpace(host)

	return host == "" || host == "127.0.0.1" || host == "localhost"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
func portAvailable(address string) bool {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

func mountAPI(r chi.Router, api *app.API) {
	r.Route("/api", func(router chi.Router) {
		router.Get("/products", handleListProducts(api))
		router.Post("/products", handleCreateProduct(api))
		router.Put("/products/{id}", handleUpdateProduct(api))
		router.Post("/products/{id}/adjust-stock", handleAdjustStock(api))
		router.Post("/products/{id}/archive", handleArchiveProduct(api))
		router.Delete("/products/{id}", handleDeleteProduct(api))

		router.Get("/customers", handleListCustomers(api))
		router.Post("/customers", handleCreateCustomer(api))
		router.Put("/customers/{id}", handleUpdateCustomer(api))
		router.Delete("/customers/{id}", handleDeleteCustomer(api))

		router.Get("/orders", handleListOrders(api))
		router.Post("/orders", handleCreateOrder(api))
		router.Delete("/orders/{id}", handleDeleteOrder(api))
		router.Post("/orders/{id}/label", handleGenerateLabel(api))
		router.Get("/orders/export.csv", handleExportOrdersCSV(api))

		router.Get("/settings", handleGetSettings(api))
		router.Put("/settings", handleUpdateSettings(api))

		router.Get("/couriers", handleListCouriers(api))
		router.Post("/couriers", handleCreateCourier(api))
		router.Put("/couriers/{id}", handleUpdateCourier(api))
		router.Delete("/couriers/{id}", handleDeleteCourier(api))

		router.Get("/stock-opnames", handleListStockOpnames(api))
		router.Post("/stock-opnames", handlePerformStockOpname(api))

		router.Post("/backups", handleCreateBackup(api))
		router.Post("/backups/restore", handleRestoreBackup(api))
	})
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("write json error: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	if err == nil {
		err = fmt.Errorf("unknown error")
	}
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func handleListProducts(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		page := parsePositiveInt(query.Get("page"), 1)
		pageSize := parsePositiveInt(query.Get("pageSize"), 20)
		if pageSize <= 0 {
			pageSize = 20
		}
		search := strings.TrimSpace(query.Get("q"))

		result, err := api.ListProducts(r.Context(), service.ProductListOptions{
			Query:    search,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func handleCreateProduct(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload domain.Product
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		payload.ID = ""
		created, err := api.SaveProduct(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func handleUpdateProduct(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var payload domain.Product
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		payload.ID = id
		updated, err := api.SaveProduct(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func handleAdjustStock(api *app.API) http.HandlerFunc {
	type request struct {
		Delta  int    `json:"delta"`
		Reason string `json:"reason"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var payload request
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if err := api.AdjustStock(r.Context(), id, payload.Delta, payload.Reason); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

func handleArchiveProduct(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := api.ArchiveProduct(r.Context(), id); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

func handleDeleteProduct(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := api.DeleteProduct(r.Context(), id); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

func handleListCustomers(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		page := parsePositiveInt(query.Get("page"), 1)
		pageSize := parsePositiveInt(query.Get("pageSize"), 20)
		if pageSize <= 0 {
			pageSize = 20
		}
		search := strings.TrimSpace(query.Get("q"))

		result, err := api.ListCustomers(r.Context(), service.CustomerListOptions{
			Query:    search,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func handleCreateCustomer(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload domain.Customer
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		payload.ID = ""
		created, err := api.SaveCustomer(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func handleUpdateCustomer(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var payload domain.Customer
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		payload.ID = id
		updated, err := api.SaveCustomer(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func handleDeleteCustomer(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := api.DeleteCustomer(r.Context(), id); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

func handleListOrders(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		page := parsePositiveInt(query.Get("page"), 1)
		pageSize := parsePositiveInt(query.Get("pageSize"), 5)
		if pageSize <= 0 {
			pageSize = 5
		}
		search := strings.TrimSpace(query.Get("q"))
		courier := strings.TrimSpace(query.Get("courier"))
		dateStartStr := strings.TrimSpace(query.Get("dateStart"))
		dateEndStr := strings.TrimSpace(query.Get("dateEnd"))

		var dateStartPtr, dateEndPtr *time.Time
		if dateStartStr != "" {
			parsed, err := time.Parse("2006-01-02", dateStartStr)
			if err != nil {
				writeError(w, http.StatusBadRequest, fmt.Errorf("tanggal mulai tidak valid"))
				return
			}
			dateStartPtr = &parsed
		}
		if dateEndStr != "" {
			parsed, err := time.Parse("2006-01-02", dateEndStr)
			if err != nil {
				writeError(w, http.StatusBadRequest, fmt.Errorf("tanggal akhir tidak valid"))
				return
			}
			dateEndPtr = &parsed
		}

		result, err := api.ListOrders(r.Context(), service.OrderListOptions{
			Query:     search,
			Courier:   courier,
			DateStart: dateStartPtr,
			DateEnd:   dateEndPtr,
			Page:      page,
			PageSize:  pageSize,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func handleCreateOrder(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload service.CreateOrderInput
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		created, err := api.CreateOrder(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func handleDeleteOrder(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := api.DeleteOrder(r.Context(), id); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

func handleGenerateLabel(api *app.API) http.HandlerFunc {
	type response struct {
		Base64 string `json:"base64"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		label, err := api.GenerateLabel(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, response{Base64: label})
	}
}

func handleGetSettings(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := api.GetSettings(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, settings)
	}
}

func handleUpdateSettings(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload domain.AppSettings
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		updated, err := api.UpdateSettings(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func handleListCouriers(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		page := parsePositiveInt(query.Get("page"), 1)
		pageSize := parsePositiveInt(query.Get("pageSize"), 20)
		if pageSize <= 0 {
			pageSize = 20
		}
		search := strings.TrimSpace(query.Get("q"))

		result, err := api.ListCouriers(r.Context(), service.CourierListOptions{
			Query:    search,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func handleCreateCourier(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload domain.Courier
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		payload.ID = ""
		created, err := api.SaveCourier(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func handleUpdateCourier(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var payload domain.Courier
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		payload.ID = id
		updated, err := api.SaveCourier(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func handleDeleteCourier(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := api.DeleteCourier(r.Context(), id); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

func handleListStockOpnames(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 10
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		opnames, err := api.ListStockOpnames(r.Context(), limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, opnames)
	}
}

func handlePerformStockOpname(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload service.PerformStockOpnameInput
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		opname, err := api.PerformStockOpname(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, opname)
	}
}

func handleCreateBackup(api *app.API) http.HandlerFunc {
	type request struct {
		IncludeSchema *bool `json:"includeSchema"`
		IncludeData   *bool `json:"includeData"`
		IncludeMedia  *bool `json:"includeMedia"`
	}
	type response struct {
		Payload string `json:"payload"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		opts := domain.BackupOptions{IncludeSchema: true, IncludeData: true, IncludeMedia: true}
		if r.Body != nil && r.ContentLength != 0 {
			var payload request
			if err := decodeJSON(r.Body, &payload); err != nil {
				if !errors.Is(err, io.EOF) {
					writeError(w, http.StatusBadRequest, err)
					return
				}
			}
			if payload.IncludeSchema != nil {
				opts.IncludeSchema = *payload.IncludeSchema
			}
			if payload.IncludeData != nil {
				opts.IncludeData = *payload.IncludeData
			}
			if payload.IncludeMedia != nil {
				opts.IncludeMedia = *payload.IncludeMedia
			}
		} else if r.Body != nil {
			_ = r.Body.Close()
		}

		backup, err := api.CreateBackup(r.Context(), opts)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, response{Payload: backup})
	}
}

func handleRestoreBackup(api *app.API) http.HandlerFunc {
	type request struct {
		Payload                 string `json:"payload"`
		DisableForeignKeyChecks *bool  `json:"disableForeignKeyChecks"`
		UseTransaction          *bool  `json:"useTransaction"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var payload request
		if err := decodeJSON(r.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		opts := domain.RestoreOptions{DisableForeignKeyChecks: true, UseTransaction: true}
		if payload.DisableForeignKeyChecks != nil {
			opts.DisableForeignKeyChecks = *payload.DisableForeignKeyChecks
		}
		if payload.UseTransaction != nil {
			opts.UseTransaction = *payload.UseTransaction
		}
		if strings.TrimSpace(payload.Payload) == "" {
			writeError(w, http.StatusBadRequest, fmt.Errorf("payload is required"))
			return
		}
		summary, err := api.RestoreBackup(r.Context(), payload.Payload, opts)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, summary)
	}
}

func handleExportOrdersCSV(api *app.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		filters := service.OrderExportFilters{
			Search:  strings.TrimSpace(query.Get("search")),
			Courier: strings.TrimSpace(query.Get("courier")),
		}

		if raw := strings.TrimSpace(query.Get("start")); raw != "" {
			if ts, err := time.Parse("2006-01-02", raw); err == nil {
				start := ts
				filters.Start = &start
			}
		}
		if raw := strings.TrimSpace(query.Get("end")); raw != "" {
			if ts, err := time.Parse("2006-01-02", raw); err == nil {
				end := ts.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
				filters.End = &end
			}
		}

		csvBytes, err := api.ExportOrdersCSV(r.Context(), filters)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		filename := fmt.Sprintf("orders-%s.csv", time.Now().UTC().Format("20060102-150405"))
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(csvBytes); err != nil {
			log.Printf("write csv response: %v", err)
		}
	}
}

func decodeJSON(body io.ReadCloser, target any) error {
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

type spaHandler struct {
	fs fs.FS
}

func newSPAHandler(dist fs.FS) http.Handler {
	return &spaHandler{fs: dist}
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}
	file, err := h.fs.Open(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			log.Printf("serve static error: %v", err)
		}
		h.serveIndex(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		h.serveIndex(w, r)
		return
	}
	if info.IsDir() {
		h.serveIndex(w, r)
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		h.serveIndex(w, r)
		return
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(data))
}

func (h *spaHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	file, err := h.fs.Open("index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(data))
}

func resolveFrontendDist() (fs.FS, string, error) {
	if dist, err := fs.Sub(embeddedDist, "frontend/dist"); err == nil {
		if hasFrontendEntry(dist) {
			return dist, "embedded bundle", nil
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, "", fmt.Errorf("load embedded dist: %w", err)
	}

	disk := os.DirFS("frontend/dist")
	if hasFrontendEntry(disk) {
		return disk, "frontend/dist directory", nil
	}

	return nil, "", fmt.Errorf("frontend build not found. jalankan `npm -C frontend install` lalu `npm -C frontend run build`")
}

func hasFrontendEntry(fsys fs.FS) bool {
	if fsys == nil {
		return false
	}
	if _, err := fs.Stat(fsys, "index.html"); err == nil {
		return true
	}
	return false
}

func newMissingFrontendHandler(err error) http.Handler {
	message := "Aset frontend belum tersedia. Jalankan `npm -C frontend install` dan `npm -C frontend run build` sebelum menjalankan server Go."
	if err != nil {
		message = fmt.Sprintf("%s\n\nDetail: %v", message, err)
	}
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, message)
	})
}
