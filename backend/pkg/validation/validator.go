// Package validation expõe um validador de structs baseado em tags
// (go-playground/validator), com mensagens de erro traduzidas e nomes de
// campo derivados da tag `json` (não do nome Go), pra bater com o que o
// client realmente enviou.
package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/cia-da-vacina/crm/backend/internal/domain/vo"
	"github.com/go-playground/validator/v10"
)

var (
	instance *validator.Validate
	once     sync.Once
)

func get() *validator.Validate {
	once.Do(func() {
		v := validator.New()

		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return fld.Name
			}
			return name
		})

		_ = v.RegisterValidation("e164", func(fl validator.FieldLevel) bool {
			return vo.IsE164(fl.Field().String())
		})

		instance = v
	})
	return instance
}

// Struct valida uma struct usando suas tags e retorna um map[campo]mensagem.
// Retorna nil se não houver erros.
func Struct(s any) map[string]string {
	err := get().Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return map[string]string{"_": "invalid input"}
	}

	errs := make(map[string]string, len(ve))
	for _, fe := range ve {
		errs[fe.Field()] = fieldMessage(fe)
	}
	return errs
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email (ex: example@example.com)"
	case "e164":
		return "must be a valid E.164 phone number (ex: +5511999998888)"
	case "min":
		return fmt.Sprintf("must have at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must have at most %s characters", fe.Param())
	case "uuid":
		return "must be a valid UUID (ex: 00000000-0000-0000-0000-000000000000)"
	case "oneof":
		return fmt.Sprintf("must be one of the following values: %s", fe.Param())
	default:
		return fmt.Sprintf("invalid (%s)", fe.Tag())
	}
}
