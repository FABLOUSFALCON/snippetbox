package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	//nolint:gosec // pprof is intentionally enabled in debug mode only
	_ "net/http/pprof"

	"github.com/FABLOUSFALCON/snippetbox/internal/models"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

/* =========================
   Configuration
   ========================= */

type config struct {
	addr     string
	dsn      string
	debug    bool
	certFile string
	keyFile  string
	useTLS   bool
}

func parseFlags() config {
	addr := flag.String("addr", ":4001", "HTTP network address")
	dsn := flag.String("dsn", "", "PostgreSQL data source name")
	debug := flag.Bool("debug", false, "Enable debug mode")
	certFile := flag.String("cert", "./tls/localhost+1.pem", "TLS certificate file path")
	keyFile := flag.String("key", "./tls/localhost+1-key.pem", "TLS key file path")
	useTLS := flag.Bool("tls", false, "Enable TLS (use false for cloud platforms like Render)")

	flag.Parse()

	// Priority: 1. Flag, 2. Env var, 3. Default
	dsnValue := *dsn
	if dsnValue == "" {
		dsnValue = os.Getenv("DATABASE_URL")
		if dsnValue == "" {
			// Fallback to default local DSN
			dsnValue = "postgres://web:pass@localhost:5432/snippetbox?sslmode=disable"
		}
	}

	return config{
		addr:     *addr,
		dsn:      dsnValue,
		debug:    *debug,
		certFile: *certFile,
		keyFile:  *keyFile,
		useTLS:   *useTLS,
	}
}

/* =========================
   Application
   ========================= */

type application struct {
	debug          bool
	logger         *slog.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	db             *pgxpool.Pool
}

/* =========================
   Entry Point
   ========================= */

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := parseFlags()

	logger := newLogger(cfg.debug)

	if cfg.debug {
		startPprof(logger)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		return err
	}

	db, err := openDB(cfg.dsn)
	if err != nil {
		return err
	}
	defer closeDB(logger, db)

	app := newApplication(cfg, logger, templateCache, db, cfg.dsn)

	srv := newHTTPServer(cfg, app, logger)

	logger.Info("starting server", slog.String("addr", cfg.addr), slog.Bool("tls", cfg.useTLS))

	// Use TLS only if configured (local dev), otherwise use plain HTTP (Render handles SSL)
	if cfg.useTLS {
		err = srv.ListenAndServeTLS(cfg.certFile, cfg.keyFile)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

/* =========================
   Logger
   ========================= */

func newLogger(debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	return slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		}),
	)
}

/* =========================
   pprof (debug only)
   ========================= */

func startPprof(logger *slog.Logger) {
	pprofSrv := &http.Server{
		Addr:         "localhost:6060",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logger.Info("pprof enabled", slog.String("addr", pprofSrv.Addr))
		if err := pprofSrv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error("pprof server error", slog.String("err", err.Error()))
		}
	}()
}

/* =========================
   Application wiring
   ========================= */

func newApplication(
	cfg config,
	logger *slog.Logger,
	templateCache map[string]*template.Template,
	db *pgxpool.Pool,
	dsn string,
) *application {
	formDecoder := form.NewDecoder()

	// Create sql.DB connection for session store (postgresstore requires it)
	sessionDB, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("failed to open session database", slog.String("err", err.Error()))
		// Fall back to memory store if DB connection fails
		sessionDB = nil
	}

	sessionManager := scs.New()
	if sessionDB != nil {
		sessionManager.Store = postgresstore.NewWithCleanupInterval(sessionDB, 30*time.Minute)
	}
	sessionManager.Lifetime = 12 * time.Hour
	// Only set secure cookies when using TLS
	sessionManager.Cookie.Secure = cfg.useTLS

	return &application{
		debug:          cfg.debug,
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		db:             db,
	}
}

/* =========================
   HTTP server
   ========================= */

func newHTTPServer(
	cfg config,
	app *application,
	logger *slog.Logger,
) *http.Server {
	return &http.Server{
		Addr:     cfg.addr,
		Handler:  app.routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
			MinVersion: tls.VersionTLS12,
		},
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

/* =========================
   Database
   ========================= */

func openDB(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse config for connection pooling
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Set pool configuration for performance
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func closeDB(logger *slog.Logger, db *pgxpool.Pool) {
	db.Close()
	logger.Info("database connection pool closed")
}
