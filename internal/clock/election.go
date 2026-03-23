package clock

import (
	"context"
	"errors"
	"time"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

const (
	leaderKey = "market:leader"

	leaseTTL       = 3 * time.Second          // how long the lock lives without renewal
	renewInterval  = leaseTTL / 3             // renew well before expiry
	retryInterval  = 500 * time.Millisecond   // how often a standby retries acquisition
)

type Election struct {
	rdb      *redis.Client
	leaderID string // unique ID for this instance (e.g. hostname + pid)
	log  *slog.Logger
}

func NewElection(rdb *redis.Client, leaderID string) *Election {
	return &Election{rdb: rdb, leaderID: leaderID}
}


// Campaign blocks until this instance wins the election, then calls
// onElected in a goroutine. When onElected returns (or ctx is cancelled),
// the lease is released and Campaign returns — callers should loop:
//
//	for {
//	    if err := e.Campaign(ctx, onElected); !errors.Is(err, ErrNotRenewed) { break }
//	}
func (e *Election) Campaign(ctx context.Context, onElected func(ctx context.Context)) error {
	// --- Phase 1: acquire the lock ---
	for {
		ok, err := e.rdb.SetNX(ctx, leaderKey, e.leaderID, leaseTTL).Result()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			e.log.Error("acquire error", "error", err)
		} else if ok {
			break // we are the leader
		}

		// Lock is held by someone else; wait before retrying.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
		}
	}

	e.log.Error("leader elected", "leader_id", e.leaderID)

	// --- Phase 2: hold the lock via a renewal goroutine ---
	leaderCtx, abdicate := context.WithCancel(ctx)
	defer abdicate()

	renewErr := make(chan error, 1)
	go func() {
		renewErr <- e.holdLease(leaderCtx)
	}()

	// Run the master work in a separate goroutine so we can watch for
	// lease loss at the same time.
	done := make(chan struct{})
	go func() {
		defer close(done)
		onElected(leaderCtx)
	}()

	var result error
	select {
	case err := <-renewErr:
		// Lease lost — tell onElected to stop.
		abdicate()
		<-done
		result = err
	case <-done:
		// onElected finished on its own (shouldn't happen in normal operation).
		result = nil
	}

	// Best-effort release so the next leader doesn't wait for TTL.
	e.release(context.Background())
	return result
}

// holdLease renews the lease until ctx is cancelled or renewal fails.
func (e *Election) holdLease(ctx context.Context) error {
	ticker := time.NewTicker(renewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Only renew if we still own the key.
			renewed, err := e.tryRenew(ctx)
			if err != nil {
				e.log.Error("renewal error", "error", err)
				return ErrNotRenewed
			}
			if !renewed {
				e.log.Error("lost leadership", "leader_id", e.leaderID)
				return ErrNotRenewed
			}
		}
	}
}

// tryRenew uses a Lua script so the check-and-set is atomic.
var renewScript = redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("PEXPIRE", KEYS[1], ARGV[2])
	else
		return 0
	end
`)

func (e *Election) tryRenew(ctx context.Context) (bool, error) {
	ttlMs := leaseTTL.Milliseconds()
	res, err := renewScript.Run(ctx, e.rdb, []string{leaderKey}, e.leaderID, ttlMs).Int()
	return res == 1, err
}

// release deletes the key only if we still own it.
var releaseScript = redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`)

func (e *Election) release(ctx context.Context) {
	if _, err := releaseScript.Run(ctx, e.rdb, []string{leaderKey}, e.leaderID).Int(); err != nil {
		e.log.Error("release error (non-fatal)", "error", err)
	}
}

// ErrNotRenewed is returned when a leader loses its lease.
var ErrNotRenewed = errors.New("leader lease not renewed")
