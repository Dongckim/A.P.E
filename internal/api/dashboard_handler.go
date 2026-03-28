package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/dongchankim/ape/internal/sftp"
)

type DashboardHandler struct {
	connMgr *ConnectionManager
}

func NewDashboardHandler(connMgr *ConnectionManager) *DashboardHandler {
	return &DashboardHandler{connMgr: connMgr}
}

func (h *DashboardHandler) getClient(r *http.Request) sftp.SFTPClient {
	connID := r.URL.Query().Get("conn")
	if connID != "" {
		return h.connMgr.Get(connID)
	}
	return h.connMgr.Default()
}

// --- Overview ---

type SystemOverview struct {
	Hostname    string     `json:"hostname"`
	OS          string     `json:"os"`
	Kernel      string     `json:"kernel"`
	Arch        string     `json:"arch"`
	UptimeSince string     `json:"uptime_since"`
	CPU         CPUInfo    `json:"cpu"`
	Memory      MemoryInfo `json:"memory"`
	Disks       []DiskInfo `json:"disks"`
}

type CPUInfo struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
}

type MemoryInfo struct {
	Total     int64 `json:"total"`
	Used      int64 `json:"used"`
	Available int64 `json:"available"`
}

type DiskInfo struct {
	Mount   string `json:"mount"`
	Total   int64  `json:"total"`
	Used    int64  `json:"used"`
	Avail   int64  `json:"avail"`
	Percent int    `json:"percent"`
}

func (h *DashboardHandler) HandleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	ctx := r.Context()
	overview := SystemOverview{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// uname
	wg.Add(1)
	go func() {
		defer wg.Done()
		out, _, err := client.Exec(ctx, "uname -a")
		if err != nil {
			return
		}
		parts := strings.Fields(strings.TrimSpace(out))
		mu.Lock()
		if len(parts) >= 1 {
			overview.Kernel = parts[0]
		}
		if len(parts) >= 2 {
			overview.Hostname = parts[1]
		}
		if len(parts) >= 3 {
			overview.OS = parts[2]
		}
		for _, p := range parts {
			if p == "x86_64" || p == "aarch64" || p == "arm64" {
				overview.Arch = p
				break
			}
		}
		mu.Unlock()
	}()

	// OS pretty name
	wg.Add(1)
	go func() {
		defer wg.Done()
		out, _, _ := client.Exec(ctx, "cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d'\"' -f2")
		if name := strings.TrimSpace(out); name != "" {
			mu.Lock()
			overview.OS = name
			mu.Unlock()
		}
	}()

	// uptime
	wg.Add(1)
	go func() {
		defer wg.Done()
		out, _, _ := client.Exec(ctx, "uptime -s 2>/dev/null")
		mu.Lock()
		overview.UptimeSince = strings.TrimSpace(out)
		mu.Unlock()
	}()

	// CPU
	wg.Add(1)
	go func() {
		defer wg.Done()
		out, _, _ := client.Exec(ctx, "nproc 2>/dev/null")
		cores, _ := strconv.Atoi(strings.TrimSpace(out))
		out2, _, _ := client.Exec(ctx, "top -bn1 | grep 'Cpu(s)' | awk '{print $2}'")
		usage, _ := strconv.ParseFloat(strings.TrimSpace(out2), 64)
		mu.Lock()
		overview.CPU = CPUInfo{UsagePercent: usage, Cores: cores}
		mu.Unlock()
	}()

	// Memory
	wg.Add(1)
	go func() {
		defer wg.Done()
		out, _, _ := client.Exec(ctx, "free -b | grep Mem")
		fields := strings.Fields(strings.TrimSpace(out))
		if len(fields) >= 7 {
			total, _ := strconv.ParseInt(fields[1], 10, 64)
			used, _ := strconv.ParseInt(fields[2], 10, 64)
			avail, _ := strconv.ParseInt(fields[6], 10, 64)
			mu.Lock()
			overview.Memory = MemoryInfo{Total: total, Used: used, Available: avail}
			mu.Unlock()
		}
	}()

	// Disk
	wg.Add(1)
	go func() {
		defer wg.Done()
		out, _, _ := client.Exec(ctx, "df -B1 --output=target,size,used,avail,pcent 2>/dev/null | tail -n +2")
		var disks []DiskInfo
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}
			// Skip tmpfs, devtmpfs, etc
			if strings.HasPrefix(fields[0], "/dev") || fields[0] == "/" {
				total, _ := strconv.ParseInt(fields[1], 10, 64)
				used, _ := strconv.ParseInt(fields[2], 10, 64)
				avail, _ := strconv.ParseInt(fields[3], 10, 64)
				pct, _ := strconv.Atoi(strings.TrimSuffix(fields[4], "%"))
				disks = append(disks, DiskInfo{
					Mount: fields[0], Total: total, Used: used, Avail: avail, Percent: pct,
				})
			}
		}
		mu.Lock()
		overview.Disks = disks
		mu.Unlock()
	}()

	wg.Wait()
	writeJSON(w, http.StatusOK, overview)
}

// --- Services ---

type ServicesData struct {
	Runtimes []string        `json:"available_runtimes"`
	Docker   *DockerData     `json:"docker"`
	PM2      *PM2Data        `json:"pm2"`
	Systemd  *SystemdData    `json:"systemd"`
}

type DockerData struct {
	Containers []DockerContainer `json:"containers"`
	Error      string            `json:"error,omitempty"`
}

type DockerContainer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	Ports  string `json:"ports"`
	State  string `json:"state"`
}

type PM2Data struct {
	Processes []PM2Process `json:"processes"`
	Error     string       `json:"error,omitempty"`
}

type PM2Process struct {
	Name     string  `json:"name"`
	ID       int     `json:"id"`
	Mode     string  `json:"mode"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu"`
	Memory   int64   `json:"memory"`
	Restarts int     `json:"restarts"`
	Uptime   int64   `json:"uptime"`
}

type SystemdData struct {
	Services []SystemdService `json:"services"`
}

type SystemdService struct {
	Unit        string `json:"unit"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
}

func (h *DashboardHandler) HandleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	ctx := r.Context()
	data := ServicesData{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Detect runtimes
	detect, _, _ := client.Exec(ctx, "which docker 2>/dev/null && echo DOCKER; which pm2 2>/dev/null && echo PM2; which systemctl 2>/dev/null && echo SYSTEMCTL")
	if strings.Contains(detect, "DOCKER") {
		data.Runtimes = append(data.Runtimes, "docker")
	}
	if strings.Contains(detect, "PM2") {
		data.Runtimes = append(data.Runtimes, "pm2")
	}
	if strings.Contains(detect, "SYSTEMCTL") {
		data.Runtimes = append(data.Runtimes, "systemd")
	}

	// Docker
	if strings.Contains(detect, "DOCKER") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, code, _ := client.Exec(ctx, `docker ps -a --format '{"id":"{{.ID}}","name":"{{.Names}}","image":"{{.Image}}","status":"{{.Status}}","ports":"{{.Ports}}","state":"{{.State}}"}' 2>&1`)
			dd := &DockerData{}
			if code != 0 {
				dd.Error = strings.TrimSpace(out)
			} else {
				for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
					if line == "" {
						continue
					}
					var c DockerContainer
					if err := json.Unmarshal([]byte(line), &c); err == nil {
						dd.Containers = append(dd.Containers, c)
					}
				}
			}
			mu.Lock()
			data.Docker = dd
			mu.Unlock()
		}()
	}

	// PM2
	if strings.Contains(detect, "PM2") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, code, _ := client.Exec(ctx, "pm2 jlist 2>&1")
			pd := &PM2Data{}
			if code != 0 {
				pd.Error = strings.TrimSpace(out)
			} else {
				var raw []map[string]any
				if err := json.Unmarshal([]byte(out), &raw); err == nil {
					for _, item := range raw {
						p := PM2Process{}
						if v, ok := item["name"].(string); ok { p.Name = v }
						if v, ok := item["pm_id"].(float64); ok { p.ID = int(v) }
						if v, ok := item["pm2_env"].(map[string]any); ok {
							if s, ok := v["exec_mode"].(string); ok { p.Mode = s }
							if s, ok := v["status"].(string); ok { p.Status = s }
							if n, ok := v["restart_time"].(float64); ok { p.Restarts = int(n) }
							if n, ok := v["pm_uptime"].(float64); ok { p.Uptime = int64(n) }
						}
						if v, ok := item["monit"].(map[string]any); ok {
							if n, ok := v["cpu"].(float64); ok { p.CPU = n }
							if n, ok := v["memory"].(float64); ok { p.Memory = int64(n) }
						}
						pd.Processes = append(pd.Processes, p)
					}
				}
			}
			mu.Lock()
			data.PM2 = pd
			mu.Unlock()
		}()
	}

	// Systemd
	if strings.Contains(detect, "SYSTEMCTL") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, _, _ := client.Exec(ctx, "systemctl list-units --type=service --state=running --no-pager --plain --no-legend 2>/dev/null")
			sd := &SystemdData{}
			for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
				fields := strings.Fields(line)
				if len(fields) < 4 {
					continue
				}
				sd.Services = append(sd.Services, SystemdService{
					Unit:        fields[0],
					Active:      fields[2],
					Sub:         fields[3],
					Description: strings.Join(fields[4:], " "),
				})
			}
			mu.Lock()
			data.Systemd = sd
			mu.Unlock()
		}()
	}

	wg.Wait()
	if data.Runtimes == nil {
		data.Runtimes = []string{}
	}
	writeJSON(w, http.StatusOK, data)
}

// --- Git ---

type GitInfo struct {
	Path       string      `json:"path"`
	Branch     string      `json:"branch"`
	LastCommit *GitCommit  `json:"last_commit"`
	Commits    []GitCommit `json:"commits"`
	Error      string      `json:"error,omitempty"`
}

type GitCommit struct {
	Hash         string `json:"hash"`
	Message      string `json:"message"`
	Author       string `json:"author"`
	RelativeDate string `json:"relative_date,omitempty"`
	Date         string `json:"date,omitempty"`
}

func (h *DashboardHandler) HandleGitLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	ctx := r.Context()
	info := GitInfo{Path: path}

	// Branch
	branchOut, code, _ := client.Exec(ctx, fmt.Sprintf("cd %s && git branch --show-current 2>&1", path))
	if code != 0 {
		info.Error = "not a git repository"
		writeJSON(w, http.StatusOK, info)
		return
	}
	info.Branch = strings.TrimSpace(branchOut)

	// Last commit
	lastOut, _, _ := client.Exec(ctx, fmt.Sprintf("cd %s && git log -1 --format='%%H|%%s|%%an|%%ai' 2>/dev/null", path))
	if parts := strings.SplitN(strings.TrimSpace(lastOut), "|", 4); len(parts) == 4 {
		info.LastCommit = &GitCommit{
			Hash: parts[0][:minLen(len(parts[0]), 7)], Message: parts[1], Author: parts[2], Date: parts[3],
		}
	}

	// Recent commits
	logOut, _, _ := client.Exec(ctx, fmt.Sprintf("cd %s && git log --format='%%H|%%s|%%an|%%ar' -20 2>/dev/null", path))
	for _, line := range strings.Split(strings.TrimSpace(logOut), "\n") {
		parts := strings.SplitN(line, "|", 4)
		if len(parts) == 4 {
			info.Commits = append(info.Commits, GitCommit{
				Hash: parts[0][:minLen(len(parts[0]), 7)], Message: parts[1], Author: parts[2], RelativeDate: parts[3],
			})
		}
	}

	writeJSON(w, http.StatusOK, info)
}

func minLen(a, b int) int {
	if a < b { return a }
	return b
}

// --- Processes ---

type ProcessInfo struct {
	PID     string  `json:"pid"`
	User    string  `json:"user"`
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Command string  `json:"command"`
}

func (h *DashboardHandler) HandleProcesses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	out, _, _ := client.Exec(r.Context(), "ps aux --sort=-%mem 2>/dev/null | head -21")
	lines := strings.Split(strings.TrimSpace(out), "\n")

	var procs []ProcessInfo
	// Skip header line
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)
		procs = append(procs, ProcessInfo{
			PID:     fields[1],
			User:    fields[0],
			CPU:     cpu,
			Memory:  mem,
			Command: strings.Join(fields[10:], " "),
		})
	}

	writeJSON(w, http.StatusOK, procs)
}
