# radar-agent

`radar-agent` is a small Go service that measures HTTP and ICMP latency from a
VPS, PC, or server and sends minute samples to a `radar-api` hub.

Each minute it chooses a random second from 5 through 50, downloads the active
targets, runs their enabled probes concurrently, and submits one batch. HTTP
counts any received response (including 4xx/5xx) as reachable. Failed and timed
out probes are submitted with `ok=false` and `latency_ms=null`.

## Configuration

Copy `.env.example` to `.env`, or provide the same values through the process
environment:

| Variable | Required | Default |
|---|---:|---|
| `RADAR_HUB_URL` | No | `http://127.0.0.1:8088` |
| `RADAR_AGENT_TOKEN` | Yes | Demo: `radar-demo-agent-token-probe1` |
| `RADAR_HTTP_TIMEOUT_SEC` | No | `5` |
| `RADAR_ICMP_TIMEOUT_SEC` | No | `3` |

The `.env` file is optional. The agent authenticates every hub request with
`Authorization: Bearer <RADAR_AGENT_TOKEN>`.

## Build and run

```powershell
go mod tidy
go build -o bin/radar-agent.exe ./cmd/radar-agent
Copy-Item .env.example .env
.\bin\radar-agent.exe
```

On Linux:

```bash
go build -o bin/radar-agent ./cmd/radar-agent
RADAR_AGENT_TOKEN=radar-demo-agent-token-probe1 ./bin/radar-agent
```

Raw ICMP commonly requires elevated privileges. When raw ICMP is unavailable,
the agent falls back to the operating system's `ping` command. Ensure `ping` is
installed and available on `PATH`.

## Linux systemd installation

1. Build or copy the binary to `/opt/radar-agent/radar-agent`.
2. Create a dedicated `radar-agent` system user and make the binary executable.
3. Copy `deploy/radar-agent.service` to `/etc/systemd/system/`.
4. Create `/etc/radar-agent.env` with the variables above and restrict it to the
   service user.
5. Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now radar-agent
sudo systemctl status radar-agent
```

If raw ICMP is desired without running as root, grant only the required Linux
capability:

```bash
sudo setcap cap_net_raw+ep /opt/radar-agent/radar-agent
```

## Windows installation

Run `radar-agent.exe` directly for an interactive process, or create a Windows
Scheduled Task that starts it at boot. Set the environment variables at the
machine or task level, or place `.env` in the task's working directory. Configure
the task to restart on failure and run whether or not a user is logged in.

Press Ctrl+C, stop the systemd service, or terminate the Scheduled Task to trigger
a graceful shutdown.
