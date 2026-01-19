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
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
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
}

func parseFlags() config {
	addr := flag.String("addr", ":4001", "HTTP network address")
	dsn := flag.String("dsn", "web:lolamancer@/snippetbox?parseTime=true", "MySQL data source name")
	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	return config{
		addr:     *addr,
		dsn:      *dsn,
		debug:    *debug,
		certFile: "./tls/localhost+1.pem",
		keyFile:  "./tls/localhost+1-key.pem",
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

	app := newApplication(cfg, logger, templateCache, db)

	srv := newHTTPServer(cfg, app, logger)

	logger.Info("starting server", slog.String("addr", cfg.addr))

	err = srv.ListenAndServeTLS(cfg.certFile, cfg.keyFile)
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
	db *sql.DB,
) *application {
	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &application{
		debug:          cfg.debug,
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
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

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func closeDB(logger *slog.Logger, db *sql.DB) {
	if err := db.Close(); err != nil {
		logger.Error("closing database", slog.String("err", err.Error()))
	}
}
