package configs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type ConsulService struct {
	ID      string            `json:"ID"`
	Name    string            `json:"Name"`
	Address string            `json:"Address"`
	Port    int               `json:"Port"`
	Check   map[string]string `json:"Check"`
}

// RegisterService registers the service with Consul
func RegisterService(serviceID, serviceName, address string, port int, healthCheckURL string) error {
	service := ConsulService{
		ID:      serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Check: map[string]string{
			"HTTP":     healthCheckURL,
			"Interval": "10s",
		},
	}

	data, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service data: %v", err)
	}

	consulAddress := os.Getenv("CONSUL_ADDRESS")
	if consulAddress == "" {
		consulAddress = "http://localhost:8500" // 기본값 설정
	}

	url := fmt.Sprintf("%s/v1/agent/service/register", consulAddress)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register service with Consul: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register service with Consul: %s", resp.Status)
	}

	log.Printf("Service '%s' registered successfully with Consul", serviceName)
	return nil
}
