package grantprovider

import (
	"encoding/json"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// GetConfigDir devuelve la ruta ~/.grant, creando el directorio si no existe.
// Puede terminar el proceso con log.Fatal si falla el home o la creación.
func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory")
	}
	path := filepath.Join(home, ".grant")
	log.Tracef("Checking if %s directory exists", path)
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatalf("Error creating config directory %s", path)
	}
	return path
}

// GetConfig lee ~/.grant/fileName en dest. Si el archivo no existe, escribe
// defaultConfig en disco y asigna dest con ese valor.
func GetConfig[T any](fileName string, dest *T, defaultConfig T) error {
	path := filepath.Join(GetConfigDir(), fileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Warnf("Config file %s not exists", path)
		data, err := json.Marshal(defaultConfig)
		if err != nil {
			return err
		}
		log.Tracef("Writing default configuration in file %s", path)
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
