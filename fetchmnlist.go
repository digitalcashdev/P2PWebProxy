package p2pwebproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type MasternodeResponse struct {
	Result map[string]MasternodeInfo `json:"result"`
}

type MasternodeInfo struct {
	ProTxHash           string `json:"proTxHash"`
	Address             string `json:"address"`
	Payee               string `json:"payee"`
	Status              string `json:"status"`
	Type                string `json:"type"`
	PlatformNodeID      string `json:"platformNodeID"`
	PlatformP2PPort     int    `json:"platformP2PPort"`
	PlatformHTTPPort    int    `json:"platformHTTPPort"`
	Pospenaltyscore     int    `json:"pospenaltyscore"`
	ConsecutivePayments int    `json:"consecutivePayments"`
	Lastpaidtime        int    `json:"lastpaidtime"`
	Lastpaidblock       int    `json:"lastpaidblock"`
	Owneraddress        string `json:"owneraddress"`
	Votingaddress       string `json:"votingaddress"`
	Collateraladdress   string `json:"collateraladdress"`
	Pubkeyoperator      string `json:"pubkeyoperator"`
}

func FetchAllowedList(baseURL, user, pass string) (map[string][]string, error) {
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
	for _, mn := range masternodeResp.Result {
		if mn.Status != "ENABLED" {
			log.Printf("skipping disabled node '%s'", mn.Address)
			continue
		}

		parts := strings.Split(mn.Address, ":")
		if len(parts) != 2 {
			log.Printf("skipping invalid address '%s': %v", mn.Address, err)
			continue
		}

		hostname := parts[0]
		port := parts[1]
		allowedList[hostname] = append(allowedList[hostname], port)
	}

	return allowedList, nil
}
