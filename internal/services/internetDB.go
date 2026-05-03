package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

const defaultInternetDBBaseURL = "https://internetdb.shodan.io"

var defaultInternetDBClient = &http.Client{Timeout: 15 * time.Second}

type InternetDBResponse struct {
	IP        string   `json:"ip"`
	Ports     []int    `json:"ports"`
	Hostnames []string `json:"hostnames"`
	Tags      []string `json:"tags"`
	Vulns     []string `json:"vulns"`
	CPEs      []string `json:"cpes"`
}

type InternetDBService struct {
	baseURL string
	client  *http.Client
}

func NewInternetDBService(baseURL string, client *http.Client) *InternetDBService {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultInternetDBBaseURL
	}
	if client == nil {
		client = defaultInternetDBClient
	}

	return &InternetDBService{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

func CallInternetDB(ctx context.Context, ip string) (*InternetDBResponse, error) {
	return NewInternetDBService("", nil).LookupByIP(ctx, ip)
}

func (s *InternetDBService) LookupByIP(ctx context.Context, ip string) (*InternetDBResponse, error) {
	parsedIP, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		return nil, fmt.Errorf("internetdb requires a valid IP address: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/"+parsedIP.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create InternetDB request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("InternetDB request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read InternetDB response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("InternetDB API error %d: %s", resp.StatusCode, string(body))
	}

	var result InternetDBResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal InternetDB response: %w", err)
	}

	return &result, nil
}
