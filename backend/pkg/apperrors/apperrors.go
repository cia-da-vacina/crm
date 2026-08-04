package apperrors

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsNotFound retorna true se o erro for sql.ErrNoRows.
// Permite que o usecase detecte "não encontrado" sem importar database/sql.
func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// IsUniqueViolation retorna true se o erro for violação de constraint UNIQUE
// do Postgres (código 23505). Permite mapear race de criação para 409.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// IsForeignKeyViolation retorna true se o erro for violação de FK do Postgres:
// foreign_key_violation (23503) ou restrict_violation (23001, FK ON DELETE RESTRICT).
func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == "23503" || pgErr.Code == "23001")
}

// ResponseError é um erro de regra de negócio — o usecase decide o status e a
// mensagem que o handler deve devolver ao client.
type ResponseError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	StatusCode int               `json:"-"`
	Details    map[string]string `json:"details,omitempty"`
}

func (e *ResponseError) Error() string {
	return e.Message
}

func NewValidationError(details map[string]string) *ResponseError {
	return &ResponseError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		StatusCode: http.StatusBadRequest,
		Details:    details,
	}
}

func NewBadRequestError(message string) *ResponseError {
	return &ResponseError{
		Code:       "BAD_REQUEST",
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func NewUnauthorizedError(message string) *ResponseError {
	if message == "" {
		message = "Unauthorized"
	}
	return &ResponseError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewForbiddenError(message string) *ResponseError {
	if message == "" {
		message = "Access denied"
	}
	return &ResponseError{
		Code:       "FORBIDDEN",
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

func NewNotFoundError(resource string) *ResponseError {
	return &ResponseError{
		Code:       "NOT_FOUND",
		Message:    resource + " not found",
		StatusCode: http.StatusNotFound,
	}
}

func NewConflictError(resource string) *ResponseError {
	return &ResponseError{
		Code:       "CONFLICT",
		Message:    resource + " already exists",
		StatusCode: http.StatusConflict,
	}
}

// NewConflictErrorMessage é como NewConflictError, mas com mensagem livre —
// usado quando o 409 não é "já existe" (ex.: conversa já reivindicada por
// outro atendente, reenvio de OTP sem pendência ativa).
func NewConflictErrorMessage(message string) *ResponseError {
	return &ResponseError{
		Code:       "CONFLICT",
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

func NewTooManyRequestsError(message string) *ResponseError {
	return &ResponseError{
		Code:       "TOO_MANY_REQUESTS",
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
	}
}

// NewBadGatewayError representa falha ao chamar um provedor externo (Meta,
// OpenAI) — o backend está de pé, mas a dependência downstream falhou.
func NewBadGatewayError(message string) *ResponseError {
	return &ResponseError{
		Code:       "BAD_GATEWAY",
		Message:    message,
		StatusCode: http.StatusBadGateway,
	}
}

// SystemError é um erro inesperado (falha de infra, bug) — sempre vira 500,
// mas carrega a causa original pra log/debug sem expô-la ao client em produção.
type SystemError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *SystemError) Error() string {
	return e.Message
}

func (e *SystemError) Unwrap() error {
	return e.Cause
}

func NewInternalError(cause error) *SystemError {
	return &SystemError{
		Code:    "INTERNAL_ERROR",
		Message: "An internal error occurred",
		Cause:   cause,
	}
}

func NewDatabaseError(cause error) *SystemError {
	return &SystemError{
		Code:    "DATABASE_ERROR",
		Message: "A database error occurred",
		Cause:   cause,
	}
}
