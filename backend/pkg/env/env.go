package env

import (
	"os"
	"strconv"
)

// GetOrDefault retorna o valor da variável de ambiente convertido para o tipo T,
// ou o valor padrão caso a variável não esteja definida ou a conversão falhe.
// Tipos suportados: string, int, int64, float64.
func GetOrDefault[T string | int | int64 | float64](key string, defaultValue T) T {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}

	var result any
	switch any(defaultValue).(type) {
	case string:
		result = raw
	case int:
		v, err := strconv.Atoi(raw)
		if err != nil {
			return defaultValue
		}
		result = v
	case int64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return defaultValue
		}
		result = v
	case float64:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return defaultValue
		}
		result = v
	default:
		return defaultValue
	}

	if r, ok := result.(T); ok {
		return r
	}
	return defaultValue
}

func IsDev() bool {
	return GetOrDefault("ENVIRONMENT", "dev") == "dev"
}
