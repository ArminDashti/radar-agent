# Description

Radar Agent is a Go CLI/service that polls a Radar API hub for monitoring
targets, measures HTTP and ICMP latency once per clock minute, and submits
authenticated sample batches. It uses the standard library, `x/net/icmp`, and
`godotenv`; the entry point is `cmd/radar-agent/main.go`.

Run with `go run ./cmd/radar-agent` after setting `RADAR_AGENT_TOKEN`.
