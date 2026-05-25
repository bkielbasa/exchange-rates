package fetcher

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/bklimczak/exchange-rates/internal/rates"
)

type ecbrssFetcher struct {
	httpClient *http.Client
	sourceURL  string
	logger     *slog.Logger
}

func NewECBRSSFetcher(sourceURL string, logger *slog.Logger) *ecbrssFetcher{
	if logger == nil {
		logger = slog.Default()
	}
	return &ecbrssFetcher{
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		sourceURL: sourceURL,
		logger:    logger,
	}
}

type rssItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type rssFeed struct {
	XMLName xml.Name  `xml:"rss"`
	Items   []rssItem `xml:"channel>item"`
}

// pairRE matches one "CODE VALUE" entry inside an item description, e.g.
// "USD 1.16480000". The feed lists every currency in a single description.
var pairRE = regexp.MustCompile(`([A-Z]{3})\s+(-?[0-9]+(?:\.[0-9]+)?)`)

func (f *ecbrssFetcher) Fetch(ctx context.Context, currency string) (rates.Rate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.sourceURL, nil)
	if err != nil {
		return rates.Rate{}, fmt.Errorf("build request: %w", err)
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return rates.Rate{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rates.Rate{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return rates.Rate{}, fmt.Errorf("read body: %w", err)
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return rates.Rate{}, fmt.Errorf("parse rss: %w", err)
	}
	if len(feed.Items) == 0 {
		return rates.Rate{}, fmt.Errorf("empty feed")
	}

	latest := pickLatestItem(feed.Items)
	want := strings.ToUpper(strings.TrimSpace(currency))
	for _, m := range pairRE.FindAllStringSubmatch(latest.Description, -1) {
		if m[1] != want {
			continue
		}
		rate, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			return rates.Rate{}, fmt.Errorf("parse rate for %s: %w", want, err)
		}
		return rates.Rate{
			Currency:  want,
			Rate:      rate,
			FetchedAt: parsePubDateOrNow(latest.PubDate),
		}, nil
	}
	return rates.Rate{}, fmt.Errorf("currency %s not found in feed", want)
}

// pickLatestItem returns the item with the most recent parseable pubDate, or
// the last item if no dates parse (the feed is chronological, oldest first).
func pickLatestItem(items []rssItem) rssItem {
	latest := items[len(items)-1]
	latestT := parsePubDateOrZero(latest.PubDate)
	for _, it := range items {
		t := parsePubDateOrZero(it.PubDate)
		if t.After(latestT) {
			latest = it
			latestT = t
		}
	}
	return latest
}

func parsePubDateOrZero(s string) time.Time {
	if t, err := time.Parse(time.RFC1123, s); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse(time.RFC1123Z, s); err == nil {
		return t.UTC()
	}
	return time.Time{}
}

func parsePubDateOrNow(s string) time.Time {
	if t := parsePubDateOrZero(s); !t.IsZero() {
		return t
	}
	return time.Now().UTC()
}
