package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type ShodanResponse struct {
	Matches []struct {
		IP        string `json:"ip_str"`
		Port      int    `json:"port"`
		Org       string `json:"org"`
		Hostnames []string `json:"hostnames"`
	} `json:"matches"`
}

func CallShodan() (ShodanResponse, error){
	apiKey := os.Getenv("SHODAN_API_KEY")
	if apiKey == "" {
		return ShodanResponse{}, fmt.Errorf("SHODAN_API_KEY environment variable is not set")
	}

	query := "apache"
	endpoint := "https://api.shodan.io/shodan/host/search"
	params := url.Values{}
	params.Add("key", apiKey)
	params.Add("query", query)

	resp, err := http.Get(endpoint + "?" + params.Encode())
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result ShodanResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		panic(err)
	}

	return result, nil
}