package middleware

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
)

func testRedisAddr() string {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	return host + ":" + port
}

// TestRateLimiter_SharedAcrossInstances is the "two replicas behind a load
// balancer" scenario: two independent stores/limiters (as main.go and
// auth/routes.go each construct their own) pointed at the same Redis must
// enforce one combined limit rather than each replica getting its own
// private quota.
func TestRateLimiter_SharedAcrossInstances(t *testing.T) {
	addr := testRedisAddr()
	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		t.Skipf("redis not reachable at %s, skipping: %v", addr, err)
	}

	const limit = int64(5)
	rate := limiter.Rate{Period: time.Minute, Limit: limit}

	storeA, err := NewLimiterStore(client, "test_rl_shared")
	if err != nil {
		t.Fatalf("failed to build store A: %v", err)
	}
	storeB, err := NewLimiterStore(client, "test_rl_shared")
	if err != nil {
		t.Fatalf("failed to build store B: %v", err)
	}

	limiterA := limiter.New(storeA, rate)
	limiterB := limiter.New(storeB, rate)

	key := fmt.Sprintf("shared-ip-%d", time.Now().UnixNano())
	defer client.Del(context.Background(), "test_rl_shared:"+key)

	var totalAllowed int64
	var lastReached bool

	// Alternate requests between the two "replicas", past the shared limit.
	for i := 0; i < int(limit)+3; i++ {
		l := limiterA
		if i%2 == 1 {
			l = limiterB
		}

		lc, err := l.Get(context.Background(), key)
		if err != nil {
			t.Fatalf("Get failed on request %d: %v", i, err)
		}
		if !lc.Reached {
			totalAllowed++
		}
		lastReached = lc.Reached
	}

	if totalAllowed != limit {
		t.Fatalf("expected exactly %d allowed requests shared across both instances, got %d", limit, totalAllowed)
	}
	if !lastReached {
		t.Fatalf("expected the limit to be reached once combined requests exceeded %d", limit)
	}
}

// TestNewLimiterStore_MemoryFallback verifies a nil Redis client falls back
// to an in-memory store instead of erroring, for local dev without Redis.
func TestNewLimiterStore_MemoryFallback(t *testing.T) {
	store, err := NewLimiterStore(nil, "test_rl_memory")
	if err != nil {
		t.Fatalf("expected no error for memory fallback, got: %v", err)
	}
	if store == nil {
		t.Fatal("expected a non-nil store")
	}
}
