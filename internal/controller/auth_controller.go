package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/service"
)

const (
	SessionCookieName = "session_id"
	SessionDuration   = 24 * time.Hour
)

type AuthController struct {
	authService service.AuthServiceInterface
	sessions    *model.SessionStore
}

func NewAuthController(authService service.AuthServiceInterface) *AuthController {
	return &AuthController{
		authService: authService,
		sessions:    model.GetSessionStore(),
	}
}

func (ac *AuthController) SetRoutes(r *mux.Router) {
	r.HandleFunc("/auth/register", ac.Register).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", ac.Login).Methods(http.MethodPost)
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"securepassword123"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"securepassword123"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User    *model.User `json:"user"`
	Message string      `json:"message,omitempty"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "Registration details"
// @Success      201      {object}  AuthResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /auth/register [post]
func (ac *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := ac.authService.CreateUser(req.Email, req.Password)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			RespondWithError(w, http.StatusConflict, "User already exists")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	RespondWithJSON(w, http.StatusCreated, AuthResponse{
		User:    user,
		Message: "User created successfully",
	})
}

// Login godoc
// @Summary      Login
// @Description  Authenticate user and create session
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Login credentials"
// @Success      200      {object}  AuthResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /auth/login [post]
func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := ac.authService.Login(req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to login")
		return
	}

	// Create session
	sessionID := ac.sessions.Create(user.ID, user.Email, SessionDuration)

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(SessionDuration.Seconds()),
	})

	RespondWithJSON(w, http.StatusOK, AuthResponse{
		User:    user,
		Message: "Login successful",
	})
}

// Helper functions (now exported so middleware can use them)
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, ErrorResponse{
		Status:  "error",
		Message: message,
	})
}
