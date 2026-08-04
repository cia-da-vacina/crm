package apperrors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// MapDBError converte erros do banco em apperrors conhecidos.
// constraints mapeia o nome da constraint (ex: "users_email_unique") para o campo legível (ex: "E-mail").
// Erros não reconhecidos são retornados sem alteração para o usecase envolver em DatabaseError.
func MapDBError(err error, constraints map[string]string) error {
	if ok, constraint := isUniqueViolation(err); ok {
		if field, found := constraints[constraint]; found {
			return NewConflictError(field)
		}
		return NewConflictError("registry")
	}
	if ok, constraint := isForeignKeyViolation(err); ok {
		if field, found := constraints[constraint]; found {
			return NewNotFoundError(field)
		}
		return NewBadRequestError("invalid reference")
	}
	return err
}

// isUniqueViolation retorna true e o nome da constraint se o erro for PostgreSQL 23505 (unique_violation).
func isUniqueViolation(err error) (bool, string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true, pgErr.ConstraintName
	}
	return false, ""
}

// isForeignKeyViolation retorna true e o nome da constraint se o erro for PostgreSQL 23503 (foreign_key_violation).
func isForeignKeyViolation(err error) (bool, string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return true, pgErr.ConstraintName
	}
	return false, ""
}
