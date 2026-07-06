package app

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
)

// DefaultEngineTickInterval is how often the engine polls for expired offers.
const DefaultEngineTickInterval = 5 * time.Second

// DispatchEngine is a background worker that auto-retries expired offers.
// When a driver fails to respond within the offer TTL, the engine clears the
// offer and immediately attempts the next nearest driver.
//
// Lifecycle guarantees:
//   - Start is idempotent: only the first call launches the background goroutine.
//   - Stop waits for all in-progress job goroutines to finish before returning.
//   - A job already being processed is skipped on the next tick (sync.Map guard).
type DispatchEngine struct {
	jobRepo      repository.DispatchJobRepository
	locationRepo repository.DriverLocationRepository
	tripUpdater  repository.TripUpdater
	tickInterval time.Duration
	radiusKM     float64
	searchLimit  int
	done         chan struct{}
	startOnce    sync.Once
	stopOnce     sync.Once
	wg           sync.WaitGroup
	inFlight     sync.Map // jobID → struct{}
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

// Start launches the background goroutine. Subsequent calls are no-ops.
func (e *DispatchEngine) Start() {
	e.startOnce.Do(func() {
		ticker := time.NewTicker(e.tickInterval)
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
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
	})
}

// Stop signals the engine to stop and waits for all in-flight job goroutines to finish.
// Subsequent calls are no-ops.
func (e *DispatchEngine) Stop() {
	e.stopOnce.Do(func() {
		close(e.done)
	})
	e.wg.Wait()
}

func (e *DispatchEngine) processExpiredOffers(ctx context.Context) {
	now := time.Now().UTC()
	jobs, err := e.jobRepo.FindExpiredOffers(ctx, now)
	if err != nil {
		log.Error().Err(err).Msg("dispatch engine: find expired offers failed")
		return
	}
	for _, job := range jobs {
		if _, loaded := e.inFlight.LoadOrStore(job.JobID, struct{}{}); loaded {
			continue // another goroutine is already processing this job
		}
		j := job
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
			defer e.inFlight.Delete(j.JobID)
			e.processJob(ctx, j, now)
		}()
	}
}

func (e *DispatchEngine) processJob(ctx context.Context, job *entity.DispatchJob, now time.Time) {
	if err := job.TimeoutOffer(now); err != nil {
		log.Warn().Err(err).Str("job_id", job.JobID).Msg("dispatch engine: timeout offer skipped")
		return
	}
	if err := offerNextDriver(ctx, job, e.locationRepo, e.tripUpdater, e.jobRepo, e.radiusKM, e.searchLimit); err != nil {
		log.Error().Err(err).Str("job_id", job.JobID).Msg("dispatch engine: offer next driver failed")
	}
}
