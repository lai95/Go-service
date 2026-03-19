package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type Config struct {
	ServerURL   string `json:"server_url"`
	IntervalSec int    `json:"interval_sec"`
	Status      string `json:"status"`
}

type Heartbeat struct {
	Hostname  string   `json:"hostname"`
	Timestamp string   `json:"timestamp"`
	Status    string   `json:"status"`
	IPs       []string `json:"ips,omitempty"`
}

func main() {
	cfg := loadConfig("/home/michael/gopro/client/config.json")

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("failed to get hostname: %v", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	log.Printf("heartbeat agent started: server=%s hostname=%s interval=%ds", cfg.ServerURL, hostname, cfg.IntervalSec)

	sendHeartbeat(client, cfg, hostname)

	ticker := time.NewTicker(time.Duration(cfg.IntervalSec) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sendHeartbeat(client, cfg, hostname)
	}
}

func loadConfig(path string) Config {
	cfg := Config{
		ServerURL:   "http://127.0.0.1:8080/heartbeat",
		IntervalSec: 10,
		Status:      "alive",
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("config not found, using defaults: %v", err)
		return cfg
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("invalid config, using defaults: %v", err)
		return cfg
	}

	if cfg.ServerURL == "" {
		cfg.ServerURL = "http://127.0.0.1:8080/heartbeat"
	}
	if cfg.IntervalSec <= 0 {
		cfg.IntervalSec = 10
	}
	if cfg.Status == "" {
		cfg.Status = "alive"
	}

	return cfg
}

func sendHeartbeat(client *http.Client, cfg Config, hostname string) {
	hb := Heartbeat{
		Hostname:  hostname,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Status:    cfg.Status,
		IPs:       getIPv4s(),
	}

	data, err := json.Marshal(hb)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, cfg.ServerURL, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("request creation error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("heartbeat failed: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("heartbeat sent: status_code=%d hostname=%s", resp.StatusCode, hostname)
}

func getIPv4s() []string {
	var ips []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP == nil {
				continue
			}

			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}

			ips = append(ips, ip.String())
		}
	}

	return ips
}
