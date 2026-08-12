# Directory tree

```text
radar-agent/
├── cmd/radar-agent/main.go       # Process entry point and signal handling
├── deploy/radar-agent.service    # Example Linux systemd service
├── docs/                         # Maintainer-facing project documentation
├── internal/config/config.go     # Environment and dotenv configuration
├── internal/hub/client.go        # Authenticated Radar API client
├── internal/loop/loop.go         # Minute scheduling, probing, and retries
├── internal/probe/http.go        # HTTP latency probe
├── internal/probe/icmp.go        # Raw ICMP and ping fallback
├── .env.example                  # Example runtime configuration
├── .gitignore                    # Local and build exclusions
├── go.mod                        # Go module declaration
├── go.sum                        # Dependency checksums
└── README.md                     # Installation and usage guide
```

Tests are colocated with their packages as `*_test.go`.
