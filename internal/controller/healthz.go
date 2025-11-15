package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
)

var (
	// ErrNotReady is an error returned when the service is not available.
	ErrNotReady = errors.New("not ready")
)

// Checker is an interface that defines a dependency service health check.
type Checker interface {
	// Check performs the health check and returns an error if the service is unhealthy.
	Check() error
}

// NewDBChecker creates a new database health check.
func NewDBChecker(db *sqlx.DB) Checker {
	return DBChecker{
		name: "DB",
		db:   db,
	}
}

// DBChecker is a health check for the database connection.
type DBChecker struct {
	name string
	db   *sqlx.DB
}

// Check performs the health check for the database connection.
func (d DBChecker) Check() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("[DB] %w", ErrNotReady)
	}

	return nil
}

// DependencyName returns the name of the database dependency.
func (d DBChecker) DependencyName() string {
	return d.name
}

// WithDBChecker is a functional option to add a database health check to the HealthCheck controller.
func WithDBChecker() HealthCheckOption {
	return WithChecker(NewDBChecker(model.DBX()))
}

// WithChecker is a functional option to add a custom health check to the HealthCheck controller.
func WithChecker(c Checker) HealthCheckOption {
	return func(h *HealthCheck) {
		h.checkers = append(h.checkers, c)
	}
}

// HealthCheckOption is a functional option to configure the HealthCheck controller.
type HealthCheckOption func(*HealthCheck)

// NewHealthCheck creates a new HealthCheck controller.
func NewHealthCheck(opts ...HealthCheckOption) *HealthCheck {
	h := &HealthCheck{}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

type HealthCheck struct {
	checkers []Checker
}

func (h HealthCheck) SetRoutes(r *mux.Router) {
	r.HandleFunc("/api/healthz", h.checkHealth).Methods(http.MethodGet)
	r.HandleFunc("/api/readyz", h.checkReady).Methods(http.MethodGet)
}

// checkHealth is a handler that checks if the service is healthy.
func (h HealthCheck) checkHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// checkReady is a handler that checks if the service is ready to handle requests.
func (h HealthCheck) checkReady(w http.ResponseWriter, r *http.Request) {
	// if the number of checkers increases, this can be further optimized to run in parallel.
	for _, c := range h.checkers {
		if err := c.Check(); errors.Is(err, ErrNotReady) {
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, err := w.Write([]byte(err.Error())); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
