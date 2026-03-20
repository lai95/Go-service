package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Heartbeat struct {
	Hostname  string   `json:"hostname"`
	Timestamp string   `json:"timestamp"`
	Status    string   `json:"status"`
	IPs       []string `json:"ips"`
	CPUUsage  float64  `json:"cpu_usage"`
}

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var hb Heartbeat
	if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("heartbeat: hostname=%s timestamp=%s status=%s IPs=%s CPUUsage=%.2f%%", hb.Hostname, hb.Timestamp, hb.Status, hb.IPs, hb.CPUUsage)
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/heartbeat", heartbeatHandler)
	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
