package probe

import (
	"context"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var pingTimePattern = regexp.MustCompile(`(?i)time[=<]\s*([0-9]+(?:[.,][0-9]+)?)\s*ms`)

func ICMP(ctx context.Context, host string, timeout time.Duration) (*float64, bool) {
	hostname := probeHostname(host)
	if hostname == "" {
		return nil, false
	}
	if latency, ok := rawICMP(ctx, hostname, timeout); ok {
		return latency, true
	}
	return commandPing(ctx, hostname, timeout)
}

func rawICMP(ctx context.Context, host string, timeout time.Duration) (*float64, bool) {
	ip, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return nil, false
	}
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, false
	}
	defer conn.Close()
	deadline := time.Now().Add(timeout)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	_ = conn.SetDeadline(deadline)

	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte("radar-agent")},
	}
	data, err := message.Marshal(nil)
	if err != nil {
		return nil, false
	}
	start := time.Now()
	if _, err := conn.WriteTo(data, ip); err != nil {
		return nil, false
	}
	reply := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(reply)
		if err != nil {
			return nil, false
		}
		parsed, err := icmp.ParseMessage(1, reply[:n])
		if err == nil && parsed.Type == ipv4.ICMPTypeEchoReply {
			latency := float64(time.Since(start).Microseconds()) / 1000
			return &latency, true
		}
	}
}

func commandPing(ctx context.Context, host string, timeout time.Duration) (*float64, bool) {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var args []string
	if runtime.GOOS == "windows" {
		args = []string{"-n", "1", "-w", strconv.Itoa(int(timeout.Milliseconds())), host}
	} else {
		seconds := int(timeout.Seconds())
		if seconds < 1 {
			seconds = 1
		}
		args = []string{"-c", "1", "-W", strconv.Itoa(seconds), host}
	}
	output, err := exec.CommandContext(pingCtx, "ping", args...).CombinedOutput()
	if err != nil {
		return nil, false
	}
	match := pingTimePattern.FindStringSubmatch(string(output))
	if len(match) != 2 {
		return nil, false
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
	if err != nil {
		return nil, false
	}
	return &value, true
}

func probeHostname(host string) string {
	value := strings.TrimSpace(host)
	if !strings.Contains(value, "://") {
		value = "//" + value
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	hostname := parsed.Hostname()
	if hostname == "" {
		return strings.Trim(value, "[]")
	}
	return hostname
}
