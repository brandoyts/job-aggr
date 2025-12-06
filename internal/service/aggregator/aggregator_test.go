package aggregator

import (
	"context"
	"errors"
	"testing"

	"github.com/brandoyts/job-aggr/internal/mocks"
	"github.com/brandoyts/job-aggr/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestNewAggregatorService tests that a new aggregator service can be created
func TestNewAggregatorService(t *testing.T) {
	scraper1 := mocks.NewJobScraper(t)
	scraper2 := mocks.NewJobScraper(t)

	service := NewAggregatorService(scraper1, scraper2)

	assert.NotNil(t, service)
	assert.Implements(t, (*AggregatorService)(nil), service)
}

// TestFetchJobsSingleScraper tests fetching jobs from a single scraper
func TestFetchJobsSingleScraper(t *testing.T) {
	scraper := mocks.NewJobScraper(t)
	jobs := []model.Job{
		{
			ID:       "1",
			Title:    "Go Developer",
			Company:  "Tech Corp",
			Location: "New York",
			Url:      "https://example.com/job/1",
			Source:   "linkedin",
		},
		{
			ID:       "2",
			Title:    "Backend Engineer",
			Company:  "StartUp Inc",
			Location: "New York",
			Url:      "https://example.com/job/2",
			Source:   "linkedin",
		},
	}

	scraper.On("Fetch", mock.Anything, "golang", "New York").Return(jobs, nil)

	service := NewAggregatorService(scraper)
	ctx := context.Background()

	result, err := service.FetchJobs(ctx, "golang", "New York")

	assert.NoError(t, err)
	assert.Equal(t, len(jobs), len(result))
	assert.Equal(t, jobs, result)
	scraper.AssertCalled(t, "Fetch", mock.Anything, "golang", "New York")
}

// TestFetchJobsMultipleScrapers tests fetching and aggregating jobs from multiple scrapers
func TestFetchJobsMultipleScrapers(t *testing.T) {
	scraper1 := mocks.NewJobScraper(t)
	scraper2 := mocks.NewJobScraper(t)

	jobsFromScraper1 := []model.Job{
		{
			ID:       "1",
			Title:    "Go Developer",
			Company:  "Tech Corp",
			Location: "New York",
			Url:      "https://example.com/job/1",
			Source:   "linkedin",
		},
	}

	jobsFromScraper2 := []model.Job{
		{
			ID:       "2",
			Title:    "Go Backend Engineer",
			Company:  "Cloud Systems",
			Location: "New York",
			Url:      "https://example.com/job/2",
			Source:   "jobstreet",
		},
		{
			ID:       "3",
			Title:    "Golang Software Engineer",
			Company:  "DevOps Co",
			Location: "New York",
			Url:      "https://example.com/job/3",
			Source:   "indeed",
		},
	}

	scraper1.On("Fetch", mock.Anything, "golang", "New York").Return(jobsFromScraper1, nil)
	scraper2.On("Fetch", mock.Anything, "golang", "New York").Return(jobsFromScraper2, nil)

	service := NewAggregatorService(scraper1, scraper2)
	ctx := context.Background()

	_, err := service.FetchJobs(ctx, "golang", "New York")

	assert.NoError(t, err)
	scraper1.AssertCalled(t, "Fetch", mock.Anything, "golang", "New York")
	scraper2.AssertCalled(t, "Fetch", mock.Anything, "golang", "New York")
}

// TestFetchJobsEmptyResults tests fetching when all scrapers return empty results
func TestFetchJobsEmptyResults(t *testing.T) {
	scraper1 := mocks.NewJobScraper(t)
	scraper2 := mocks.NewJobScraper(t)

	scraper1.On("Fetch", mock.Anything, "obscurequery", "location").Return([]model.Job{}, nil)
	scraper2.On("Fetch", mock.Anything, "obscurequery", "location").Return([]model.Job{}, nil)

	service := NewAggregatorService(scraper1, scraper2)
	ctx := context.Background()

	result, err := service.FetchJobs(ctx, "obscurequery", "location")

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Empty(t, result)
}

// TestFetchJobsScraperError tests that an error from one scraper stops the fetch
func TestFetchJobsScraperError(t *testing.T) {
	scraper1 := mocks.NewJobScraper(t)
	scraper2 := mocks.NewJobScraper(t)

	blockScraper1 := make(chan struct{})

	// scraper1 is blocked until after aggregator returns
	scraper1.On("Fetch", mock.Anything, "golang", "location").
		Run(func(args mock.Arguments) {
			<-blockScraper1
		}).
		Return([]model.Job{
			{ID: "1", Title: "Go Dev", Company: "Corp1", Source: "linkedin"},
		}, nil)

	// scraper2 returns error immediately
	scraper2.On("Fetch", mock.Anything, "golang", "location").
		Return(nil, errors.New("network error"))

	service := NewAggregatorService(scraper1, scraper2)
	ctx := context.Background()

	result, err := service.FetchJobs(ctx, "golang", "location")

	// Release scraper1 after aggregator returns (cleanup)
	close(blockScraper1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "network error", err.Error())

	scraper1.AssertCalled(t, "Fetch", mock.Anything, "golang", "location")
	scraper2.AssertCalled(t, "Fetch", mock.Anything, "golang", "location")
}

// TestFetchJobsNoScrapers tests fetching with no scrapers registered
func TestFetchJobsNoScrapers(t *testing.T) {
	service := NewAggregatorService()
	ctx := context.Background()

	result, err := service.FetchJobs(ctx, "golang", "location")

	assert.NoError(t, err)
	assert.Nil(t, result)
}

// TestFetchJobsContextCancellation tests behavior when context is cancelled
func TestFetchJobsContextCancellation(t *testing.T) {
	scraper := mocks.NewJobScraper(t)

	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context as soon as scraper.Fetch is called
	scraper.On("Fetch", mock.Anything, "golang", "location").
		Run(func(args mock.Arguments) {
			cancel() // cancel context immediately
		}).
		Return(nil, context.Canceled)

	service := NewAggregatorService(scraper)

	result, err := service.FetchJobs(ctx, "golang", "location")

	// Use assert.ErrorIs to safely compare context errors
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)

	scraper.AssertCalled(t, "Fetch", mock.Anything, "golang", "location")
}

// TestFetchJobsWithJobsContainingAllFields tests that all job fields are preserved
func TestFetchJobsWithJobsContainingAllFields(t *testing.T) {
	scraper := mocks.NewJobScraper(t)
	jobs := []model.Job{
		{
			ID:          "job-123",
			Title:       "Senior Go Developer",
			Company:     "Tech Innovations",
			Location:    "Remote",
			Url:         "https://careers.example.com/jobs/123",
			Source:      "linkedin",
			Salary:      "$150k - $200k",
			Description: "We are looking for an experienced Go developer...",
		},
	}

	scraper.On("Fetch", mock.Anything, "senior golang", "location").Return(jobs, nil)

	service := NewAggregatorService(scraper)
	ctx := context.Background()

	result, err := service.FetchJobs(ctx, "senior golang", "location")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))

	job := result[0]
	assert.Equal(t, "job-123", job.ID)
	assert.Equal(t, "Senior Go Developer", job.Title)
	assert.Equal(t, "Tech Innovations", job.Company)
	assert.Equal(t, "Remote", job.Location)
	assert.Equal(t, "https://careers.example.com/jobs/123", job.Url)
	assert.Equal(t, "linkedin", job.Source)
	assert.Equal(t, "$150k - $200k", job.Salary)
	assert.Equal(t, "We are looking for an experienced Go developer...", job.Description)
}

// TestFetchJobsWithDifferentQueries tests that different queries pass through
func TestFetchJobsWithDifferentQueries(t *testing.T) {
	scraper := mocks.NewJobScraper(t)
	jobs := []model.Job{{ID: "1", Title: "Python Dev", Company: "Corp", Source: "source"}}

	scraper.On("Fetch", mock.Anything, "python", "location").Return(jobs, nil)

	service := NewAggregatorService(scraper)
	ctx := context.Background()

	result, err := service.FetchJobs(ctx, "python", "location")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	scraper.AssertCalled(t, "Fetch", mock.Anything, "python", "location")
}

// TestFetchJobsLargeResultSet tests aggregating 200 jobs
func TestFetchJobsLargeResultSet(t *testing.T) {
	scraper1 := mocks.NewJobScraper(t)
	scraper2 := mocks.NewJobScraper(t)

	jobs1 := make([]model.Job, 100)
	jobs2 := make([]model.Job, 100)

	for i := 0; i < 100; i++ {
		jobs1[i] = model.Job{ID: string(rune(i + 1)), Title: "Job", Company: "Corp1", Source: "source1"}
		jobs2[i] = model.Job{ID: string(rune(i + 101)), Title: "Job", Company: "Corp2", Source: "source2"}
	}

	scraper1.On("Fetch", mock.Anything, "test", "location").Return(jobs1, nil)
	scraper2.On("Fetch", mock.Anything, "test", "location").Return(jobs2, nil)

	service := NewAggregatorService(scraper1, scraper2)
	ctx := context.Background()

	result, err := service.FetchJobs(ctx, "test", "location")

	assert.NoError(t, err)
	assert.Equal(t, 200, len(result))
}
