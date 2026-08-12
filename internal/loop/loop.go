package loop

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/ArminDashti/radar-agent/internal/hub"
	"github.com/ArminDashti/radar-agent/internal/probe"
)

type Runner struct {
	client      *hub.Client
	httpTimeout time.Duration
	icmpTimeout time.Duration
	rng         *rand.Rand
}

func New(client *hub.Client, httpTimeout, icmpTimeout time.Duration) *Runner {
	return &Runner{
		client: client, httpTimeout: httpTimeout, icmpTimeout: icmpTimeout,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *Runner) Run(ctx context.Context) error {
	for {
		next := nextProbeTime(time.Now(), r.rng)
		log.Printf("next probe at %s", next.Format(time.RFC3339))
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}

		var targets []hub.Target
		if err := retry(ctx, "fetch targets", func() error {
			var err error
			targets, err = r.client.Targets(ctx)
			return err
		}); err != nil {
			return nil
		}
		samples := r.probeAll(ctx, targets, time.Now().UTC().Truncate(time.Minute))
		if len(samples) == 0 {
			log.Printf("no enabled target protocols")
			continue
		}
		if err := retry(ctx, "submit samples", func() error {
			return r.client.Submit(ctx, samples)
		}); err != nil {
			return nil
		}
		log.Printf("submitted %d samples", len(samples))
	}
}

func (r *Runner) probeAll(ctx context.Context, targets []hub.Target, observedAt time.Time) []hub.Sample {
	samples := make(chan hub.Sample)
	var wg sync.WaitGroup
	for _, target := range targets {
		for _, protocol := range target.Protocols {
			target, protocol := target, protocol
			if protocol != "http" && protocol != "icmp" {
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				var latency *float64
				var ok bool
				if protocol == "http" {
					latency, ok = probe.HTTP(ctx, target.Host, r.httpTimeout)
				} else {
					latency, ok = probe.ICMP(ctx, target.Host, r.icmpTimeout)
				}
				samples <- hub.Sample{
					EndpointID: target.ID, Protocol: protocol,
					ObservedAt: observedAt, LatencyMS: latency, OK: ok,
				}
			}()
		}
	}
	go func() {
		wg.Wait()
		close(samples)
	}()
	var result []hub.Sample
	for sample := range samples {
		result = append(result, sample)
	}
	return result
}

func nextProbeTime(now time.Time, rng *rand.Rand) time.Time {
	second := 5 + rng.Intn(46)
	candidate := now.Truncate(time.Minute).Add(time.Duration(second) * time.Second)
	if !candidate.After(now) {
		second = 5 + rng.Intn(46)
		candidate = now.Truncate(time.Minute).Add(time.Minute + time.Duration(second)*time.Second)
	}
	return candidate
}

func retry(ctx context.Context, operation string, fn func() error) error {
	for attempt := 0; ; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			delay := retryDelay(attempt)
			log.Printf("%s failed: %v; retrying in %s", operation, err, delay)
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
}

func retryDelay(attempt int) time.Duration {
	if attempt >= 5 {
		return 30 * time.Second
	}
	return time.Second * time.Duration(1<<attempt)
}
