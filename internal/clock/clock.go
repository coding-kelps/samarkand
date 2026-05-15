package clock

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// the election leader Redis lock
	leaderLockName = "samarkand:leader"
	// the Redis key of the clock key-value store
	clockKeyName = "samarkand:clock"
	// the clock Redis PUB/SUB channel
	clockChannelName = "samarkand:clock:ticks"
	// how often is the clock updated
	tickInterval = 100 * time.Millisecond
	// how long the lock lives without renewal
	leaseTTL = 60 * time.Second
	// renew well before expiry
	renewInterval = leaseTTL / 3
	// how often a standby retries acquisition
	retryInterval = 500 * time.Millisecond
)

type MarketClock struct {
	rdb      *redis.Client
	election *Election
	value    time.Time
	lock     sync.RWMutex
	logger   *slog.Logger
}

type MarketClockConfig struct {
	Redis  RedisConfig
	Logger *slog.Logger
}

type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

func NewMarketClock(cfg *MarketClockConfig) (*MarketClock, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	logger := cfg.Logger

	e, err := NewElection(&ElectionConfig{
		Rdb:           rdb,
		Lock:          leaderLockName,
		LeaseTTL:      leaseTTL,
		RenewInterval: renewInterval,
		RetryInterval: retryInterval,
		Logger:        logger,
	})
	if err != nil {
		return nil, &ErrElectionError{err: err}
	}

	return &MarketClock{
		rdb:      rdb,
		election: e,
		logger:   logger,
	}, nil
}

func (c *MarketClock) Start(ctx context.Context) error {
	go func() {
		for {
			err := c.election.Campaign(ctx, c.send)
			if _, ok := err.(*ErrNotRenewed); !ok {
				break
			}
		}
	}()

	go c.receive(ctx)

	return nil
}

func (c *MarketClock) Close() error {
	if err := c.rdb.Close(); err != nil {
		return err
	}

	return nil
}

func (c *MarketClock) send(ctx context.Context) error {
	var clock time.Time

	ticker := time.NewTicker(tickInterval)
	res, err := c.rdb.Get(ctx, clockKeyName).Result()
	if err == redis.Nil {
		clock = time.Now() // Replace time.Now() with configured starting time.
		serialzedClock, err := clock.MarshalBinary()
		if err != nil {
			return err
		}
		if err := c.rdb.Set(ctx, clockKeyName, serialzedClock, 0).Err(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if err := clock.UnmarshalBinary([]byte(res)); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _ = <-ticker.C:
			clock = clock.Add(tickInterval)
			serialzedClock, err := clock.MarshalBinary()

			if err = c.rdb.Set(ctx, clockKeyName, serialzedClock, 0).Err(); err != nil {
				return err
			}
			if err = c.rdb.Publish(ctx, clockChannelName, serialzedClock).Err(); err != nil {
				return err
			}
		}
	}
}

func (c *MarketClock) receive(ctx context.Context) error {
	pubsub := c.rdb.Subscribe(ctx, clockChannelName)
	defer pubsub.Close()
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m := <-ch:
			c.lock.Lock()
			if err := c.value.UnmarshalBinary([]byte(m.Payload)); err != nil {
				return err
			}
			c.lock.Unlock()
		}
	}
}

func (c *MarketClock) Now() time.Time {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.value
}
