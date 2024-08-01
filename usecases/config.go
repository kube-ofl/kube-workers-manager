package usecases

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	filePath = "/etc/config/config.json"
)

type Config struct {
	ManagerPort int `json:"managerPort"`
}

func ReadConfig() (*Config, error) {
	byteData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	var config Config

	err = json.Unmarshal(byteData, &config)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file: %v\n", err)
	}

	return &config, nil
}
