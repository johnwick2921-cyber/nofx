package agent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"nofx/safe"
	"strings"
	"sync"
	"time"
)

// Brain handles proactive intelligence: the news scan. The crypto-era
// signal handling (fed by the removed Sentinel) and the market briefs (an
// external crypto exchange ticker) are gone — both had no source other than
// that external feed.
type Brain struct {
	agent    *Agent
	logger   *slog.Logger
	http     *http.Client
	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewBrain(agent *Agent, logger *slog.Logger) *Brain {
	return &Brain{
		agent:  agent,
		logger: logger,
		http:   &http.Client{Timeout: 15 * time.Second},
		stopCh: make(chan struct{}),
	}
}

func (b *Brain) Stop() {
	b.stopOnce.Do(func() {
		close(b.stopCh)
	})
}

func (b *Brain) StartNewsScan(interval time.Duration) {
	seen := make(map[string]bool)
	seenOrder := make([]string, 0, 1024)
	safe.GoNamed("brain-news-scan", func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-b.stopCh:
				return
			case <-ticker.C:
				b.scanNews(seen, &seenOrder)
			}
		}
	})
}

func (b *Brain) scanNews(seen map[string]bool, seenOrder *[]string) {
	resp, err := b.http.Get("https://min-api.cryptocompare.com/data/v2/news/?lang=EN&sortOrder=latest")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b.logger.Debug("news API non-200", "status", resp.StatusCode)
		return
	}
	body, err := safe.ReadAllLimited(resp.Body, 1024*1024) // 1MB limit
	if err != nil {
		return
	}

	var result struct {
		Data []struct {
			Title       string `json:"title"`
			Source      string `json:"source"`
			URL         string `json:"url"`
			Body        string `json:"body"`
			Categories  string `json:"categories"`
			PublishedOn int64  `json:"published_on"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return
	}

	bullish := []string{"surge", "rally", "bullish", "breakout", "ath", "pump", "adoption"}
	bearish := []string{"crash", "dump", "bearish", "sell-off", "plunge", "hack", "ban", "fraud"}

	for _, d := range result.Data {
		if seen[d.URL] {
			continue
		}
		seen[d.URL] = true
		*seenOrder = append(*seenOrder, d.URL)
		if time.Since(time.Unix(d.PublishedOn, 0)) > 10*time.Minute {
			continue
		}

		lower := strings.ToLower(d.Title + " " + d.Body)
		bc, brc := 0, 0
		for _, w := range bullish {
			if strings.Contains(lower, w) {
				bc++
			}
		}
		for _, w := range bearish {
			if strings.Contains(lower, w) {
				brc++
			}
		}

		if bc == 0 && brc == 0 {
			continue
		}

		emoji := "📰"
		sentiment := "NEUTRAL"
		if bc > brc {
			emoji = "🟢"
			sentiment = "BULLISH"
		}
		if brc > bc {
			emoji = "🔴"
			sentiment = "BEARISH"
		}

		b.agent.notifyAll(fmt.Sprintf("%s *News*\n\n%s\n\n• Source: %s\n• Sentiment: %s",
			emoji, d.Title, d.Source, sentiment))
	}

	// Evict the oldest half when seen grows large so recent URLs stay deduped deterministically.
	if len(seen) > 1000 {
		half := len(seen) / 2
		for i := 0; i < half && i < len(*seenOrder); i++ {
			delete(seen, (*seenOrder)[i])
		}
		if half < len(*seenOrder) {
			*seenOrder = append((*seenOrder)[:0], (*seenOrder)[half:]...)
		} else {
			*seenOrder = (*seenOrder)[:0]
		}
	}
}
