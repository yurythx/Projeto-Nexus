package application

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/search"
)

type fakeProvider struct {
	module string
	score  float64
	err    error
	delay  time.Duration
}

func (f fakeProvider) Module() string { return f.module }

func (f fakeProvider) Search(ctx context.Context, _ auth.Identity, q string, _ int) ([]search.Result, error) {
	select {
	case <-time.After(f.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if f.err != nil {
		return nil, f.err
	}
	return []search.Result{{Module: f.module, Title: q, Score: f.score}}, nil
}

func TestSearchDegradesGracefully(t *testing.T) {
	providers := []search.Provider{
		fakeProvider{module: "blog", score: 0.2},
		fakeProvider{module: "wiki", score: 0.9},
		fakeProvider{module: "files", err: errors.New("boom")},
		fakeProvider{module: "lento", delay: 5 * time.Second},
	}
	svc := NewService(func() []search.Provider { return providers }, slog.New(slog.NewTextHandler(io.Discard, nil)))
	svc.timeout = 100 * time.Millisecond

	resp, err := svc.Search(context.Background(), auth.Identity{}, "relatório", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 2 || resp.Results[0].Module != "wiki" {
		t.Fatalf("esperado wiki antes de blog (score), veio %+v", resp.Results)
	}
	if len(resp.Degraded) != 2 || resp.Degraded[0] != "files" || resp.Degraded[1] != "lento" {
		t.Fatalf("módulos degradados esperados [files lento], veio %v", resp.Degraded)
	}
}

func TestSearchRequiresMinimumQuery(t *testing.T) {
	svc := NewService(func() []search.Provider { return nil }, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if _, err := svc.Search(context.Background(), auth.Identity{}, "a", "", 10); err == nil {
		t.Fatal("consulta de 1 caractere deveria ser recusada")
	}
}
