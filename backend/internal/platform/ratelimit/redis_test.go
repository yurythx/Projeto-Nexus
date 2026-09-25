package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()}), mr
}

func TestRedisLimiterFixedWindow(t *testing.T) {
	client, _ := newTestRedis(t)
	l := NewRedisLimiter(client, 60, 3, "test")
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		ok, err := l.Allow(ctx, "10.0.0.1")
		if err != nil || !ok {
			t.Fatalf("requisição %d deveria passar (ok=%v err=%v)", i, ok, err)
		}
	}
	if ok, _ := l.Allow(ctx, "10.0.0.1"); ok {
		t.Fatal("4ª requisição na janela deveria ser barrada")
	}
	if ok, _ := l.Allow(ctx, "10.0.0.2"); !ok {
		t.Fatal("outra chave não pode ser afetada")
	}
}

func TestRedisLimiterAdaptivePenalty(t *testing.T) {
	client, _ := newTestRedis(t)
	l := NewRedisLimiter(client, 60, 8, "adaptive")
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := l.Penalize(ctx, "ip"); err != nil {
			t.Fatal(err)
		}
	}
	allowed := 0
	for i := 0; i < 8; i++ {
		if ok, _ := l.Allow(ctx, "ip"); ok {
			allowed++
		}
	}
	if allowed != 4 {
		t.Fatalf("com penalidade 3 o teto deveria cair de 8 para 4, passaram %d", allowed)
	}
}

func TestEffectiveLimit(t *testing.T) {
	cases := []struct{ max, penalty, want int }{
		{10, 0, 10}, {10, 2, 10}, {10, 3, 5}, {10, 6, 2}, {10, 30, 1},
	}
	for _, c := range cases {
		if got := EffectiveLimit(c.max, c.penalty); got != c.want {
			t.Errorf("EffectiveLimit(%d,%d)=%d, want %d", c.max, c.penalty, got, c.want)
		}
	}
}

func TestProgressiveDelay(t *testing.T) {
	base, max := time.Minute, time.Hour
	cases := []struct {
		n    int
		want time.Duration
	}{
		{4, 0}, {5, time.Minute}, {6, 2 * time.Minute}, {7, 4 * time.Minute}, {12, time.Hour}, {99, time.Hour},
	}
	for _, c := range cases {
		if got := ProgressiveDelay(c.n, 5, base, max); got != c.want {
			t.Errorf("ProgressiveDelay(%d)=%v, want %v", c.n, got, c.want)
		}
	}
}

func TestLockoutLifecycle(t *testing.T) {
	client, _ := newTestRedis(t)
	lo := NewLockout(client, 2, time.Minute, time.Hour)
	ctx := context.Background()

	if d, _ := lo.RegisterFailure(ctx, "user:ana"); d != 0 {
		t.Fatalf("1ª falha não bloqueia, veio %v", d)
	}
	d, err := lo.RegisterFailure(ctx, "user:ana")
	if err != nil || d != time.Minute {
		t.Fatalf("2ª falha deveria bloquear 1min (d=%v err=%v)", d, err)
	}
	if left, _ := lo.LockedFor(ctx, "user:ana"); left <= 0 {
		t.Fatal("conta deveria estar bloqueada")
	}
	if err := lo.Reset(ctx, "user:ana"); err != nil {
		t.Fatal(err)
	}
	if left, _ := lo.LockedFor(ctx, "user:ana"); left != 0 {
		t.Fatal("reset deveria liberar a conta")
	}
}
