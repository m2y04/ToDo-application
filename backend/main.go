package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appdb "todo-backend/src/db"
	"todo-backend/src/handlers"
)

type config struct {
	port           string
	databaseURL    string
	frontendOrigin string
	jwtSecret      string
}

type application struct {
	config  config
	logger  *log.Logger
	handler *handlers.Handler
}

func main() {
	cfg := config{
		port:           getEnv("PORT", "5000"),
		databaseURL:    getEnv("DATABASE_URL", "DATABASE_URL=postgres://da_user:Kc9E6ds8@localhost:5433/da_db"),
		frontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
		jwtSecret:      os.Getenv("JWT_SECRET"),
	}

	if cfg.jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	db, err := appdb.Open(cfg.databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := appdb.Ping(db); err != nil {
		log.Printf("database ping failed on startup: %v", err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config:  cfg,
		logger:  logger,
		handler: handlers.New(db, logger, cfg.jwtSecret),
	}

	server := &http.Server{
		Addr:         ":" + cfg.port,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- server.Shutdown(ctx)
	}()

	app.logger.Printf("backend running on port %s", cfg.port)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		app.logger.Fatal(err)
	}

	if err := <-shutdownError; err != nil {
		app.logger.Fatal(err)
	}

	app.logger.Println("backend stopped")
}

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", app.handler.Health)
	mux.HandleFunc("/auth/register", app.handler.Register)
	mux.HandleFunc("/auth/login", app.handler.Login)
	mux.HandleFunc("/auth/me", app.handler.Me)
	mux.HandleFunc("/todos", app.handler.Todos)
	mux.HandleFunc("/todos/", app.handler.TodoByID)

	return app.cors(mux)
}

func (app *application) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", app.config.frontendOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
