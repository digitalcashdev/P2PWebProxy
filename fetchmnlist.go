package p2pwebproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type MasternodeResponse map[string]struct {
	Address          string `json:"address"`
	Status           string `json:"status"`
	PlatformP2PPort  int    `json:"platformP2PPort"`
	PlatformHTTPPort int    `json:"platformHTTPPort"`
}

func fetchAllowedList(baseURL, user, pass string) (map[string][]string, error) {
	reqBody := []byte(`{"method": "masternodelist", "params": []}`)
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if len(user) > 0 {
		req.SetBasicAuth(user, pass)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %s, body: %s", resp.Status, string(body))
	}

	var masternodeResp MasternodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&masternodeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	allowedList := make(map[string][]string)
	for _, node := range masternodeResp {
		if node.Status == "ENABLED" {
			hostname, port, err := splitAddress(node.Address)
			if err != nil {
				log.Printf("skipping invalid address %s: %v", node.Address, err)
				continue
			}
			allowedList[hostname] = append(allowedList[hostname], port)
		}
	}

	return allowedList, nil
}

func splitAddress(address string) (string, string, error) {
	var hostname, port string
	_, err := fmt.Sscanf(address, "%s:%s", &hostname, &port)
	if err != nil {
		return "", "", fmt.Errorf("invalid address format: %v", err)
	}
	return hostname, port, nil
}
