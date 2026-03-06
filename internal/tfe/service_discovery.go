package tfe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const discoveryPath = "/.well-known/terraform.json"

// ServiceDiscovery handles service endpoint discovery
type ServiceDiscovery struct {
	serviceMap map[string]any
}

// GetServiceURL returns the URL for a service type
func (s *ServiceDiscovery) GetServiceURL(serviceType string) (string, error) {
	val, ok := s.serviceMap[serviceType]
	if !ok {
		return "", fmt.Errorf("service not found for type %s", serviceType)
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("value for service type %s is not a string", serviceType)
	}

	return str, nil
}

// DiscoverTFEServices creates a new service discovery helper by fetching the discovery document
func DiscoverTFEServices(ctx context.Context, httpClient *http.Client, endpoint string) (*ServiceDiscovery, error) {
	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	discoveryURL := url.URL{
		Scheme: parsedEndpoint.Scheme,
		Host:   parsedEndpoint.Host,
		Path:   discoveryPath,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get service discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get service discovery document, status code: %d", resp.StatusCode)
	}

	discoveredBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read service discovery document body: %w", err)
	}

	var discoveredServices map[string]any
	if err = json.Unmarshal(discoveredBytes, &discoveredServices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service discovery document body: %w", err)
	}

	return &ServiceDiscovery{serviceMap: discoveredServices}, nil
}
