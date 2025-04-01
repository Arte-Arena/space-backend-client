package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ENV_PORT             = "PORT"
	ENV_MONGODB_URI      = "MONGODB_URI"
	ACCESS_TOKEN_SECRET  = "ACCESS_TOKEN_SECRET"
	REFRESH_TOKEN_SECRET = "REFRESH_TOKEN_SECRET"
)

var allowedKeys = [2]string{ENV_PORT, ENV_MONGODB_URI}

func LoadEnvVariables() {
	workDir, err := os.Getwd()
	if err != nil {
		panic("[ENV] Erro ao obter o diretório de trabalho: " + err.Error())
	}

	filePath := filepath.Join(workDir, ".env")

	file, err := os.Open(filePath)
	if err != nil {
		panic("[ENV] Erro ao abrir o arquivo .env: " + err.Error())
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		panic("[ENV] Erro ao obter informações do arquivo .env: " + err.Error())
	}

	if fileInfo.Size() == 0 {
		panic("[ENV] O arquivo .env está vazio")
	}

	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		panic("[ENV] Erro ao criar scanner para o arquivo .env: " + err.Error())
	}

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			panic(fmt.Sprintf("[ENV] Formato inválido na linha %d: %s", lineNum, line))
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		isAllowed := false
		for _, allowedKey := range allowedKeys {
			if key == allowedKey {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			panic(fmt.Sprintf("[ENV] Chave '%s' não é permitida. Chaves permitidas: %s",
				key, strings.Join(allowedKeys[:], ", ")))
		}

		if len(value) > 1 && (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		if err := os.Setenv(key, value); err != nil {
			panic("[ENV] Erro ao definir variável de ambiente " + key + ": " + err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		panic("[ENV] Erro ao ler o arquivo .env: " + err.Error())
	}
}
