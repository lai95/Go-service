package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	CPUUsage  float64  `json:"cpu_usage"`
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

	log.Printf("heartbeat agent started: server=%s hostname=%s interval=%d", cfg.ServerURL, hostname, cfg.IntervalSec)

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
		CPUUsage:  getCPUUsage(),
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

	log.Printf("heartbeat sent: status_code=%d hostname=%s ips=%v cpu_usage=%.2f%%", resp.StatusCode, hostname, hb.IPs, hb.CPUUsage)
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
func getCPUUsage() float64 {
	idle1, total1, err := readCPUStat()
	if err != nil {
		log.Printf("failed to read cpu stat (first read): %v", err)
		return 0
	}

	time.Sleep(500 * time.Millisecond)

	idle2, total2, err := readCPUStat()
	if err != nil {
		log.Printf("failed to read cpu stat (second read): %v", err)
		return 0
	}

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)

	if totalDelta <= 0 {
		return 0
	}

	usage := 100 * (1.0 - (idleDelta / totalDelta))
	if usage < 0 {
		return 0
	}
	if usage > 100 {
		return 100
	}

	return usage
}

func readCPUStat() (idle uint64, total uint64, err error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}

	lines := bytes.Split(data, []byte("\n"))
	if len(lines) == 0 {
		return 0, 0, os.ErrInvalid
	}

	fields := bytes.Fields(lines[0])
	if len(fields) < 8 || string(fields[0]) != "cpu" {
		return 0, 0, os.ErrInvalid
	}

	var values []uint64
	for _, f := range fields[1:] {
		var v uint64
		_, err := fmt.Sscanf(string(f), "%d", &v)
		if err != nil {
			return 0, 0, err
		}
		values = append(values, v)
	}

	for _, v := range values {
		total += v
	}

	// idle + iowait
	idle = values[3]
	if len(values) > 4 {
		idle += values[4]
	}

	return idle, total, nil
}
