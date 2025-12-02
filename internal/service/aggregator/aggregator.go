package aggregator

import (
	"context"

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
	var allJobs []model.Job
	for _, scraper := range a.scrapers {
		jobs, err := scraper.Fetch(ctx, query, location)
		if err != nil {
			return nil, err
		}
		allJobs = append(allJobs, jobs...)
	}
	return allJobs, nil
}
