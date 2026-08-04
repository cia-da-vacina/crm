package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"

	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/validation"
)

func ParseJSONBody(r *http.Request, inputStruct any) error {
	return json.NewDecoder(r.Body).Decode(inputStruct)
}

// UserAgent retorna o User-Agent da request.
func UserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// ClientIP retorna o IP real do cliente, respeitando headers de proxy reverso.
func ClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.SplitN(ip, ",", 2)[0])
	}
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return strings.TrimSpace(ip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ParseAndValidate decodifica o body JSON e valida a struct.
// JSON malformado ou body vazio → BadRequestError.
// Tipo errado em campos → ValidationError com todos os campos inválidos.
// Falha nas tags validate → ValidationError com map[campo]mensagem.
// O caller pode passar o erro diretamente para http.Handle.
func ParseAndValidate(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return apperrors.NewBadRequestError("invalid request body")
	}
	if len(body) == 0 {
		return apperrors.NewBadRequestError("request body is required")
	}

	typeErrs, decodeErr := decodeCollecting(body, v)
	if decodeErr != nil {
		return apperrors.NewBadRequestError("invalid request body")
	}
	if len(typeErrs) > 0 {
		return apperrors.NewValidationError(typeErrs)
	}

	if errs := validation.Struct(v); errs != nil {
		return apperrors.NewValidationError(errs)
	}

	return nil
}

// ParseAndValidateOptional é como ParseAndValidate, mas aceita body vazio:
// nesse caso não altera v e retorna nil. Útil para PATCH com campos opcionais.
func ParseAndValidateOptional(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return apperrors.NewBadRequestError("invalid request body")
	}
	if len(body) == 0 {
		return nil
	}

	typeErrs, decodeErr := decodeCollecting(body, v)
	if decodeErr != nil {
		return apperrors.NewBadRequestError("invalid request body")
	}
	if len(typeErrs) > 0 {
		return apperrors.NewValidationError(typeErrs)
	}
	if errs := validation.Struct(v); errs != nil {
		return apperrors.NewValidationError(errs)
	}

	return nil
}

// decodeCollecting parseia o JSON em um map de valores brutos e tenta
// desserializar campo a campo, coletando todos os erros de tipo de uma vez.
// Retorna (nil, err) em caso de JSON inválido, (errs, nil) em caso de tipos errados.
func decodeCollecting(body []byte, v any) (map[string]string, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return nil, json.Unmarshal(body, v)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	typeErrs := make(map[string]string)
	re := rv.Elem()
	rt := re.Type()

	for i := range rt.NumField() {
		field := rt.Field(i)
		jsonTag := field.Tag.Get("json")
		fieldName := strings.SplitN(jsonTag, ",", 2)[0]
		if fieldName == "" || fieldName == "-" {
			fieldName = field.Name
		}

		rawVal, ok := raw[fieldName]
		if !ok {
			continue
		}

		ptr := reflect.New(field.Type)
		if err := json.Unmarshal(rawVal, ptr.Interface()); err != nil {
			var typeErr *json.UnmarshalTypeError
			if errors.As(err, &typeErr) {
				typeErrs[fieldName] = fmt.Sprintf("must be a %s", kindLabel(field.Type.Kind()))
			} else {
				typeErrs[fieldName] = "invalid value"
			}
		} else {
			re.Field(i).Set(ptr.Elem())
		}
	}

	return typeErrs, nil
}

func kindLabel(k reflect.Kind) string {
	switch k {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	default:
		return k.String()
	}
}
