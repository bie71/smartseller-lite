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
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"smartseller-lite-starter/internal/app"
	"smartseller-lite-starter/internal/db"
	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/media"
	"smartseller-lite-starter/internal/service"
)

//go:embed frontend/dist/*
var embeddedDist embed.FS

func main() {
	_ = godotenv.Load()

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
		defaultDSN = "smartseller:smartseller@tcp(127.0.0.1:3306)/smartseller"
	}
	flag.StringVar(&dsn, "dsn", defaultDSN, "MySQL DSN (e.g. user:pass@tcp(127.0.0.1:3306)/smartseller)")
	flag.Parse()

	store, err := db.NewStore(strings.TrimSpace(dsn))
	if err != nil {
		log.Fatalf("init database: %v", err)
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

	log.Printf("SmartSeller Lite listening on http://%s (db: %s, media: %s)\n", addr, store.RedactedDSN(), mediaManager.MediaDir())

	if openBrowserFlag {
		go func() {
			time.Sleep(300 * time.Millisecond)
			openDefaultBrowser(fmt.Sprintf("http://%s/", addr))
		}()
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

func openDefaultBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("failed to open browser: %v", err)
	}
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

func mountAPI(r chi.Router, api *app.API) {
	r.Route("/api", func(router chi.Router) {
		router.Get("/products", handleListProducts(api))
		router.Post("/products", handleCreateProduct(api))
		router.Put("/products/{id}", handleUpdateProduct(api))
		router.Post("/products/{id}/adjust-stock", handleAdjustStock(api))
		router.Post("/products/{id}/archive", handleArchiveProduct(api))

		router.Get("/customers", handleListCustomers(api))
		router.Post("/customers", handleCreateCustomer(api))
		router.Put("/customers/{id}", handleUpdateCustomer(api))

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
