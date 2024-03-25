package service_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

type ServiceInfo struct {
	AccessToken  string `json:"accessToken"`
	Type         string `json:"type"`
	Service      string `json:"service,omitempty"`
	Organization string `json:"organization,omitempty"`
	Location     string `json:"location,omitempty"`
}

func ParseServiceConfig(path string) (string, map[string]map[string]ServiceInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("can't open file %s, err: %s", path, err))
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("failed to close file due to err: %s\n", err)
		}
	}(file)

	dec := json.NewDecoder(file)

	var services []ServiceInfo
	err = dec.Decode(&services)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("can't decode json: %s", err))
	}

	var secretToken string
	serviceConfig := make(map[string]map[string]ServiceInfo)

	for _, service := range services {
		if service.Service == "" && (service.Type == "any" || service.Type == "to") {
			secretToken = service.AccessToken
		}
		if service.Service != "" && (service.Type == "any" || service.Type == "from") {
			if serviceConfig[service.Service] == nil {
				serviceConfig[service.Service] = make(map[string]ServiceInfo)
			}
			serviceConfig[service.Service][service.Organization] = service
		}
	}

	return secretToken, serviceConfig, nil
}
