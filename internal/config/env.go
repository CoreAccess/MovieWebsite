package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnv reads a .env file and sets environment variables.
// If the file does not exist, it returns nil.
func LoadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		// If file doesn't exist, just return
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present (only if they match at both ends)
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

