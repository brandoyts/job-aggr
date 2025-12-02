package linkedin

import (
	"context"

	"github.com/brandoyts/job-aggr/internal/model"
)

type Scraper struct{}

func NewScraper() *Scraper {
	return &Scraper{}
}

func (s *Scraper) Fetch(ctx context.Context, query string, location string) ([]model.Job, error) {
	return Fetch(ctx, query, location)
}
