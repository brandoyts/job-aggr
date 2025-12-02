package linkedin

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/brandoyts/job-aggr/internal/model"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

const (
	linkedinSearchURL = "https://www.linkedin.com/jobs/search/?keywords=%s&location=%s"
)

func Fetch(ctx context.Context, query string, location string) ([]model.Job, error) {
	path, _ := launcher.LookPath()
	u := launcher.New().
		Headless(true).  // required for backend
		NoSandbox(true). // needed for Docker and some Linux servers
		Bin(path).       // auto-detect chrome/chromium
		MustLaunch()

	browser := rod.New().
		ControlURL(u).
		Timeout(60 * time.Second).
		MustConnect()

	defer browser.MustClose()

	page := stealth.MustPage(browser) // anti-bot cloak

	parsedJob := url.QueryEscape(query)
	parsedLocation := url.QueryEscape(location)

	url := fmt.Sprintf(linkedinSearchURL, parsedJob, parsedLocation)

	page.MustNavigate(url).MustWaitLoad().MustWaitIdle()

	jobCards, err := page.Elements("div.job-search-card")
	if err != nil {
		return nil, err
	}

	var jobs []model.Job

	for _, card := range jobCards {
		title := card.MustElement("h3").MustText()

		company := ""
		if c, err := card.Element("a.hidden-nested-link"); err == nil {
			company = c.MustText()
		}

		url := ""
		if l, err := card.Element("a.base-card__full-link"); err == nil {
			link, _ := l.Attribute("href")
			url = *link
		}

		location := ""
		if c, err := card.Element("span.job-search-card__location"); err == nil {
			location = c.MustText()
		}

		jobs = append(jobs, model.Job{
			Title:    title,
			Company:  company,
			Location: location,
			Url:      url,
			Source:   "LinkedIn",
		})
	}

	return jobs, nil
}
