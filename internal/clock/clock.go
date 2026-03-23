package clock

import (
	"time"
)

const (
	clockChannel = "samarkand:clock"
	tickInterval = 100 * time.Millisecond
)

// Tick is the payload published on every clock advance.
type Tick struct {
	Seq       uint64    `json:"seq"`        // monotonic tick counter
	MarketTime time.Time `json:"market_time"` // virtual market time
	PublishedAt time.Time `json:"published_at"` // wall time of publication
}
