package domain

import (
	"testing"
	"time"
)

func TestMatchPattern(t *testing.T) {
	cases := []struct {
		pattern, key string
		want         bool
	}{
		{"#", "blog.post.published", true},
		{"blog.#", "blog.post.published", true},
		{"blog.*", "blog.post.published", false},
		{"blog.*.published", "blog.post.published", true},
		{"*.post.*", "blog.post.published", true},
		{"tramite.processo.#", "tramite.processo", true},
		{"signum.envelope.completed", "signum.envelope.completed", true},
		{"signum.envelope.completed", "signum.envelope.refused", false},
		{"#.published", "catalog.service.published", true},
		{"contact.#", "blog.post.published", false},
	}
	for _, c := range cases {
		if got := MatchPattern(c.pattern, c.key); got != c.want {
			t.Errorf("MatchPattern(%q, %q) = %v, want %v", c.pattern, c.key, got, c.want)
		}
	}
}

func TestBackoff(t *testing.T) {
	if Backoff(1) != 15*time.Second || Backoff(2) != 30*time.Second || Backoff(3) != time.Minute {
		t.Fatal("backoff exponencial inesperado")
	}
	if Backoff(50) != 6*time.Hour {
		t.Fatal("backoff deveria ter teto de 6h")
	}
}
