package env

import (
	"bufio"
	"os"
	"strings"
)

// LoadFile lê um arquivo no formato KEY=VALUE e popula as variáveis de ambiente
// que ainda não estão definidas no processo. Retorna nil silenciosamente se o
// arquivo não existir — útil em containers onde o .env pode estar ausente.
//
// Suporta:
//   - linhas em branco e comentários iniciados com #
//   - prefixo opcional "export "
//   - valores entre aspas simples ou duplas (as aspas são removidas)
//
// Variáveis já presentes no ambiente NÃO são sobrescritas, então o ambiente
// real do processo (ex.: vars do CI, do docker run -e) tem prioridade.
func LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])

		if len(val) >= 2 {
			first, last := val[0], val[len(val)-1]
			if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, val)
	}
	return scanner.Err()
}
