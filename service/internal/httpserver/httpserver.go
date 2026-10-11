package httpserver

import (
	"context"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

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
	mux := http.NewServeMux()

	iamDB, err := openAndMigrateIAM()
	if err != nil {
		log.Fatalf("IAM database: %v", err)
	}
	defer iamDB.Close()

	taskDB := db.GetNewDatabaseConnection()
	api := clientapi.GetNewClientAPI(taskDB, iamDB)

	ctx := context.Background()
	if err := api.BootstrapIAM(ctx); err != nil {
		log.Fatalf("IAM bootstrap: %v", err)
	}

	authLayer := auth.NewLayer(iamDB, runtimeconfig.RuntimeConfig.Auth, log.StandardLogger())

	apiPath, apiHandler := api.GetNewHandler(authLayer)
	apiHandler = http.StripPrefix("/api", apiHandler)
	apiHandler = withCors(apiHandler)

	mux.HandleFunc("/api"+apiPath, func(w http.ResponseWriter, r *http.Request) {
		apiHandler.ServeHTTP(w, r)
	})
	mux.Handle("/openapi", apidocs.OpenAPIHandler())
	mux.Handle("/llms.txt", apidocs.LLMsTxtHandler())
	mux.Handle("/mcp", authLayer.WrapMCPHandler(mcpserver.New(api).HTTPHandler()))
	mux.Handle("/", http.StripPrefix("/", frontend.GetNewHandler()))
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/wallpapers", getWallpapersHandler())

	endpoint := runtimeconfig.RuntimeConfig.ListenAddress

	log.WithField("endpoint", endpoint).Info("Starting HTTP server")
	log.WithField("url", endpoint+"/api"+apiPath).Info("API available")
	log.WithField("iam_db", runtimeconfig.IAMDatabasePath()).Info("IAM database ready")

	if err := http.ListenAndServe(
		endpoint,
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		log.Fatalf("Failed to start server: %v", err)
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
