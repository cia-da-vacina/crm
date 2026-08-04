package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/env"
)

type Response struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	Errors     any    `json:"errors,omitempty"`
	Metadata   any    `json:"metadata"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"success":false,"status_code":500,"message":"encoding error"}`, 500)
	}
}

func Success(w http.ResponseWriter, statusCode int, message string, data any) {
	JSON(w, statusCode, Response{
		Success:    true,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

func Created(w http.ResponseWriter, data any, message string) {
	JSON(w, http.StatusCreated, Response{
		Success:    true,
		StatusCode: http.StatusCreated,
		Message:    message,
		Data:       data,
	})
}

func Error(w http.ResponseWriter, statusCode int, message string, errs any) {
	JSON(w, statusCode, Response{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
		Errors:     errs,
	})
}

func BadRequest(w http.ResponseWriter, message string, errs any) {
	Error(w, http.StatusBadRequest, message, errs)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message, nil)
}

func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message, nil)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, nil)
}

func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, message, nil)
}

func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message, nil)
}

// Handle converte um error de aplicação na resposta HTTP correta.
// ResponseError → status + mensagem da regra de negócio.
// SystemError / qualquer outro → 500. Sempre loga o erro original.
// Em ENVIRONMENT=dev o erro original aparece no campo "errors" da resposta.
func Handle(w http.ResponseWriter, r *http.Request, err error) {
	var respErr *apperrors.ResponseError
	if errors.As(err, &respErr) {
		JSON(w, respErr.StatusCode, Response{
			Success:    false,
			StatusCode: respErr.StatusCode,
			Code:       respErr.Code,
			Message:    respErr.Message,
			Errors:     respErr.Details,
		})
		return
	}

	// Desembrulha SystemError para expor a causa raiz no log.
	cause := err
	var sysErr *apperrors.SystemError
	if errors.As(err, &sysErr) && sysErr.Cause != nil {
		cause = sysErr.Cause
	}

	code := "UNKNOWN"
	if errors.As(err, &sysErr) {
		code = sysErr.Code
	}

	log.Printf("%s %s -> 500 [%s]: %v", r.Method, r.URL.Path, code, cause)

	debug := ""
	if env.IsDev() {
		debug = cause.Error()
	}

	JSON(w, http.StatusInternalServerError, Response{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Message:    "An error occurred",
		Errors:     debug,
	})
}
