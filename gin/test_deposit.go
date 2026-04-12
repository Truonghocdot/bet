package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	applyRequest := map[string]interface{}{
		"provider":        "sepay",
		"provider_status": "finished",
		"client_ref":      "DEP-7b29b89eeeb483cdf41ed9cf8a8d84a8", 
		"provider_txn_id": "TEST12345",
		"amount":          "50000",
		"currency":        "VND",
		"paid_at":         time.Now(),
		"raw": map[string]interface{}{
			"test": "data",
		},
	}

	body, _ := json.Marshal(applyRequest)
	req, _ := http.NewRequest("POST", "http://localhost:8081/internal/v1/deposits/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", "CHANGE_THIS_INTERNAL_TOKEN")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(respBody))
}
