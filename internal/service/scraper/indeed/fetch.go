package indeed

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/brandoyts/job-aggr/internal/model"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

const (
	indeedBaseUrl   = "https://www.indeed.com"
	indeedSearchURL = "https://www.indeed.com/jobs?q=%s&l=%s"
)

func Fetch(ctx context.Context, query string, location string) ([]model.Job, error) {
	path, _ := launcher.LookPath()
	u := launcher.New().
		Headless(true).
		NoSandbox(true).
		Bin(path).
		MustLaunch()

	browser := rod.New().
		ControlURL(u).
		Timeout(60 * time.Second).
		MustConnect()

	defer browser.MustClose()

	page := stealth.MustPage(browser)

	parsedJob := url.QueryEscape(query)
	parsedLocation := url.QueryEscape(location)

	url := fmt.Sprintf(indeedSearchURL, parsedJob, parsedLocation)

	page.MustNavigate(url).MustWaitLoad().MustWaitIdle()

	jobCards, err := page.Elements("div.cardOutline")
	if err != nil {
		log.Fatal("Failed selecting job cards:", err)
		return nil, err
	}

	var jobs []model.Job

	for _, card := range jobCards {
		title := card.MustElement("h2").MustText()

		company := ""
		if c, err := card.Element(`span[data-testid="company-name"]`); err == nil {
			company = c.MustText()
		}

		url := ""
		if l, err := card.Element("a"); err == nil {
			link, _ := l.Attribute("href")
			url = *link
		}

		jobs = append(jobs, model.Job{
			Title:   title,
			Company: company,
			Url:     fmt.Sprintf("%s%s", indeedBaseUrl, url),
			Source:  "Indeed",
		})
	}

	return jobs, nil
}
