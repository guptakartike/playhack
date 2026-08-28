package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"playhack/internal/middleware"
	"playhack/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RequestOTPPayload struct {
	Email string `json:"email"`
}

type RequestOTPResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type VerifyOTPPayload struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerifyOTPResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *AuthHandler) HandleRequestOTP(w http.ResponseWriter, r *http.Request) {
	var req RequestOTPPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	code, err := h.authService.RequestOTP(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEmailDomain) {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		log.Printf("Error requesting OTP for %s: %v", req.Email, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	log.Printf("[SERVER LOG] OTP generated for %s: %s", req.Email, code)
	writeJSON(w, http.StatusOK, RequestOTPResponse{
		Message: "OTP sent successfully",
		Code:    code,
	})
}

func (h *AuthHandler) HandleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	tokenStr, err := h.authService.VerifyOTP(r.Context(), req.Email, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOTPAlreadyUsed),
			errors.Is(err, service.ErrOTPExpired),
			errors.Is(err, service.ErrOTPInvalid),
			errors.Is(err, service.ErrOTPNotFound):
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		default:
			log.Printf("Error verifying OTP for %s: %v", req.Email, err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, VerifyOTPResponse{Token: tokenStr})
}

func (h *AuthHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
