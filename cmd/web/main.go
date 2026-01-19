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

	"github.com/FABLOUSFALCON/snippetbox/internal/models"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	debug          bool
	logger         *slog.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func run() error {
	addr := flag.String("addr", ":4001", "HTTP network address")
	dsn := flag.String("dsn", "web:lolamancer@/snippetbox?parseTime=true", "MySQL data source name")
	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	// Adding our custom logger using log/slog package.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		// AddSource: true,]
	}))

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())

		return err
	}

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())

		return err
	}

	defer func() {
		err := db.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := application{
		debug:          *debug,
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:       tls.VersionTLS12,
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	// Print a log message to say that the server is starting.
	logger.Info("Starting server", slog.String("addr", *addr))

	err = srv.ListenAndServeTLS("./tls/localhost+1.pem", "./tls/localhost+1-key.pem")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		_ = db.Close()

		return nil, err
	}

	return db, nil
}
