package grantprovider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GetConfigDir devuelve la ruta ~/.grant, creando el directorio si no existe.
// Retorna error si no puede obtener el home o crear el directorio.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}
	path := filepath.Join(home, ".grant")
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("error creating config directory %s: %w", path, err)
	}
	return path, nil
}

// GetConfig lee ~/.grant/fileName en dest. Si el archivo no existe, escribe
// defaultConfig en disco y asigna dest con ese valor.
func GetConfig[T any](fileName string, dest *T, defaultConfig T) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	path := filepath.Join(configDir, fileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		data, err := json.Marshal(defaultConfig)
		if err != nil {
			return err
		}
		if err = os.WriteFile(path, data, os.ModePerm); err != nil {
			return err
		}
		*dest = defaultConfig
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
