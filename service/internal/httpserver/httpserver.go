package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	cors "github.com/rs/cors"
	log "github.com/sirupsen/logrus"

	"github.com/wacky-tracky/wacky-tracky-server/internal/apidocs"
	"github.com/wacky-tracky/wacky-tracky-server/internal/auth"
	"github.com/wacky-tracky/wacky-tracky-server/internal/clientapi"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db"
	"github.com/wacky-tracky/wacky-tracky-server/internal/frontend"
	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
	"github.com/wacky-tracky/wacky-tracky-server/internal/mcpserver"
	"github.com/wacky-tracky/wacky-tracky-server/internal/migrate"
	"github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

func StartServer() {
	if err := runServer(); err != nil {
		log.Fatalf("%v", err)
	}
}

func runServer() error {
	iamDB, err := openAndMigrateIAM()
	if err != nil {
		return fmt.Errorf("IAM database: %w", err)
	}
	defer func() {
		if cerr := iamDB.Close(); cerr != nil {
			log.WithError(cerr).Warn("Closing IAM database")
		}
	}()

	api := clientapi.GetNewClientAPI(db.GetNewDatabaseConnection(), iamDB)
	if err := api.BootstrapIAM(context.Background()); err != nil {
		return fmt.Errorf("IAM bootstrap: %w", err)
	}

	authLayer := auth.NewLayer(iamDB, runtimeconfig.RuntimeConfig.Auth, log.StandardLogger())
	apiPath, apiHandler := api.GetNewHandler(authLayer, auth.NewSSOOptOut(runtimeconfig.RuntimeConfig.Auth))
	mux := newServeMux(api, authLayer, apiPath, apiHandler)

	endpoint := runtimeconfig.RuntimeConfig.ListenAddress
	log.WithField("endpoint", endpoint).Info("Starting HTTP server")
	log.WithField("url", endpoint+"/api"+apiPath).Info("API available")
	log.WithField("iam_db", runtimeconfig.IAMDatabasePath()).Info("IAM database ready")

	if err := newHTTPServer(endpoint, mux).ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func newServeMux(api mcpserver.ClientService, authLayer *auth.Layer, apiPath string, apiHandler http.Handler) *http.ServeMux {
	apiHandler = withCors(http.StripPrefix("/api", apiHandler))

	mux := http.NewServeMux()
	mux.Handle("/api"+apiPath, apiHandler)
	mux.Handle("/openapi", apidocs.OpenAPIHandler())
	mux.Handle("/llms.txt", apidocs.LLMsTxtHandler())
	mux.Handle("/mcp", authLayer.WrapMCPHandler(mcpserver.New(api).HTTPHandler()))
	mux.Handle("/", http.StripPrefix("/", frontend.GetNewHandler()))
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/wallpapers", getWallpapersHandler())
	return mux
}

// newHTTPServer serves HTTP/1.1 and unencrypted HTTP/2 (h2c), which Connect
// clients and gRPC tooling use without TLS. No write timeout is set because
// MCP streams and git sync requests can legitimately run for a long time.
func newHTTPServer(endpoint string, handler http.Handler) *http.Server {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	return &http.Server{
		Addr:              endpoint,
		Handler:           handler,
		Protocols:         protocols,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
}

func openAndMigrateIAM() (*iamstore.SQLite, error) {
	dbPath := runtimeconfig.IAMDatabasePath()
	st, err := iamstore.Open(dbPath)
	if err != nil {
		return nil, err
	}
	if err := runIAMMigrations(st, dbPath); err != nil {
		_ = st.Close()
		return nil, err
	}
	return st, nil
}

func runIAMMigrations(st *iamstore.SQLite, dbPath string) error {
	migrationsDir, err := migrate.FindSQLiteMigrationsDir()
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"db_path":        dbPath,
		"migrations_dir": migrationsDir,
	}).Info("running IAM database migrations")

	applied, err := migrate.RunSQLiteUp(st, migrationsDir)
	if err != nil {
		return err
	}
	if applied > 0 {
		log.WithFields(log.Fields{"db_path": dbPath, "applied": applied}).Info("IAM migrations applied")
	}
	return migrate.AssertRequired(context.Background(), st, runtimeconfig.RequiredMigration, dbPath)
}

func findWallpaperDir() string {
	return "/var/www/html/wallpapers/"
}

func getWallpapersHandler() http.Handler {
	wallpaperDir := findWallpaperDir()

	log.WithFields(log.Fields{
		"wallpaperDir": wallpaperDir,
	}).Info("Starting wallpaper server")

	return http.StripPrefix("/wallpapers/", http.FileServer(http.Dir(wallpaperDir)))
}

func withCors(h http.Handler) http.Handler {
	return cors.AllowAll().Handler(h)
}
