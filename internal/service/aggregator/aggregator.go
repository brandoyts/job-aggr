package aggregator

import (
	"context"
	"sync"

	"github.com/brandoyts/job-aggr/internal/model"
)

// AggregatorService defines the interface for aggregating jobs from multiple sources.
//
//go:generate mockery --name=AggregatorService --output=../../mocks --outpkg=mocks --filename=aggregator_service_mock.go
type AggregatorService interface {
	FetchJobs(ctx context.Context, query string, location string) ([]model.Job, error)
}

// JobScraper defines the interface for fetching jobs from a specific source.
//
//go:generate mockery --name=JobScraper --output=../../mocks --outpkg=mocks --filename=job_scraper_mock.go
type JobScraper interface {
	Fetch(ctx context.Context, query string, location string) ([]model.Job, error)
}

// aggregatorService implements AggregatorService by combining results from multiple scrapers.
type aggregatorService struct {
	scrapers []JobScraper
}

// NewAggregatorService creates a new aggregator service with the given scrapers.
// Scrapers are called in order, and any error from a scraper will stop the aggregation.
func NewAggregatorService(scrapers ...JobScraper) AggregatorService {
	return &aggregatorService{scrapers: scrapers}
}

// FetchJobs fetches jobs from all registered scrapers and aggregates the results.
// If any scraper returns an error, the aggregation stops and the error is returned.
// Results are aggregated in the order that scrapers are called.
func (a *aggregatorService) FetchJobs(ctx context.Context, query string, location string) ([]model.Job, error) {
	type result struct {
		jobs []model.Job
		err  error
	}

	resultCh := make(chan result, len(a.scrapers))
	var wg sync.WaitGroup

	wg.Add(len(a.scrapers))

	for _, scraper := range a.scrapers {
		s := scraper
		go func() {
			defer wg.Done()
			jobs, err := s.Fetch(ctx, query, location)

			// Always send result, even if context is canceled
			select {
			case resultCh <- result{jobs: jobs, err: err}:
			default: // channel full or closed, skip
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var allJobs []model.Job
	for res := range resultCh {
		if res.err != nil {
			return nil, res.err
		}
		allJobs = append(allJobs, res.jobs...)
	}

	// If context was canceled before any result, return ctx.Err()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return allJobs, nil
}
