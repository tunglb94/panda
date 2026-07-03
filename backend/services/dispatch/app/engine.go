package app

import (
	"context"
	"time"

	"github.com/fairride/dispatch/domain/repository"
)

// DefaultEngineTickInterval is how often the engine polls for expired offers.
const DefaultEngineTickInterval = 5 * time.Second

// DispatchEngine is a background worker that auto-retries expired offers.
// When a driver fails to respond within the offer TTL, the engine clears the
// offer and immediately attempts the next nearest driver.
type DispatchEngine struct {
	jobRepo      repository.DispatchJobRepository
	locationRepo repository.DriverLocationRepository
	tripUpdater  repository.TripUpdater
	tickInterval time.Duration
	radiusKM     float64
	searchLimit  int
	done         chan struct{}
}

// NewDispatchEngine creates an engine with default tick interval and search params.
func NewDispatchEngine(
	jobRepo repository.DispatchJobRepository,
	locationRepo repository.DriverLocationRepository,
	tripUpdater repository.TripUpdater,
) *DispatchEngine {
	return &DispatchEngine{
		jobRepo:      jobRepo,
		locationRepo: locationRepo,
		tripUpdater:  tripUpdater,
		tickInterval: DefaultEngineTickInterval,
		radiusKM:     DefaultSearchRadiusKM,
		searchLimit:  DefaultSearchLimit,
		done:         make(chan struct{}),
	}
}

// WithTickInterval overrides the poll interval (useful for tests).
func (e *DispatchEngine) WithTickInterval(d time.Duration) *DispatchEngine {
	e.tickInterval = d
	return e
}

// Start launches the background goroutine. Call Stop to shut it down.
func (e *DispatchEngine) Start() {
	ticker := time.NewTicker(e.tickInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				e.processExpiredOffers(context.Background())
			case <-e.done:
				return
			}
		}
	}()
}

// Stop signals the engine to stop. Returns immediately; the goroutine exits
// on the next tick cycle.
func (e *DispatchEngine) Stop() {
	close(e.done)
}

func (e *DispatchEngine) processExpiredOffers(ctx context.Context) {
	now := time.Now().UTC()
	jobs, err := e.jobRepo.FindExpiredOffers(ctx, now)
	if err != nil {
		return
	}
	for _, job := range jobs {
		if err := job.TimeoutOffer(now); err != nil {
			continue
		}
		// Best-effort retry; errors are logged by the infra layer.
		_ = offerNextDriver(ctx, job, e.locationRepo, e.tripUpdater, e.jobRepo, e.radiusKM, e.searchLimit)
	}
}
