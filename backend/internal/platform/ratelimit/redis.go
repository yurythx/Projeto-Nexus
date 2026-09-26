// Package ratelimit implementa a limitação de taxa compartilhada entre as
// réplicas da API (janela fixa no Redis, adaptativa por penalidade) e o
// bloqueio progressivo de credenciais (A07).
package ratelimit

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/yurythx/projeto-nexus/internal/platform/redisx"
)

// fixedWindowScript incrementa o contador da janela e define a expiração
// na PRIMEIRA requisição da janela — atômico no Redis, sem corrida entre
// réplicas. Retorna o valor após o incremento.
var fixedWindowScript = redis.NewScript(`
local n = redis.call("INCR", KEYS[1])
if n == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return n
`)

// RedisLimiter é um limitador de janela fixa distribuído (compartilhado
// por todas as réplicas da API) que implementa httpserver.Limiter.
//
// É adaptativo (A07): cada chave acumula uma penalidade (ver Penalize)
// quando o chamador provoca falhas de autenticação/abuso, e o teto
// efetivo da janela cai pela metade a cada nível de penalidade — um IP
// que erra senhas em série recebe cada vez menos tentativas, sem afetar
// quem se comporta bem.
type RedisLimiter struct {
	client      *redis.Client
	bucket      string
	window      time.Duration
	maxRequests int
}

// NewRedisLimiter cria um limitador de até maxRequests por janela de
// windowSeconds, isolado das demais instâncias pelo nome do bucket.
func NewRedisLimiter(client *redis.Client, windowSeconds, maxRequests int, bucket string) *RedisLimiter {
	if windowSeconds <= 0 {
		windowSeconds = 60
	}
	if maxRequests <= 0 {
		maxRequests = 1
	}
	return &RedisLimiter{
		client:      client,
		bucket:      bucket,
		window:      time.Duration(windowSeconds) * time.Second,
		maxRequests: maxRequests,
	}
}

// Allow implementa httpserver.Limiter.
func (l *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	penalty, err := l.penalty(ctx, key)
	if err != nil {
		return false, err
	}
	limit := EffectiveLimit(l.maxRequests, penalty)

	windowStart := time.Now().UnixMilli() / l.window.Milliseconds()
	redisKey := redisx.Key("rl", l.bucket, key, strconv.FormatInt(windowStart, 10))
	n, err := fixedWindowScript.Run(ctx, l.client, []string{redisKey}, l.window.Milliseconds()).Int()
	if err != nil {
		return false, fmt.Errorf("ratelimit: redis incr: %w", err)
	}
	return n <= limit, nil
}

// Penalize registra uma falha para key (ex.: senha errada), elevando a
// penalidade que reduz o teto efetivo desta chave por 15 minutos.
func (l *RedisLimiter) Penalize(ctx context.Context, key string) error {
	k := redisx.Key("rl", l.bucket, "penalty", key)
	pipe := l.client.TxPipeline()
	pipe.Incr(ctx, k)
	pipe.Expire(ctx, k, 15*time.Minute)
	_, err := pipe.Exec(ctx)
	return err
}

// Forgive zera a penalidade de key (ex.: após login bem-sucedido).
func (l *RedisLimiter) Forgive(ctx context.Context, key string) error {
	return l.client.Del(ctx, redisx.Key("rl", l.bucket, "penalty", key)).Err()
}

func (l *RedisLimiter) penalty(ctx context.Context, key string) (int, error) {
	v, err := l.client.Get(ctx, redisx.Key("rl", l.bucket, "penalty", key)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("ratelimit: redis penalty: %w", err)
	}
	return v, nil
}

// EffectiveLimit reduz max pela metade a cada 3 falhas registradas, sem
// nunca descer abaixo de 1 requisição por janela.
func EffectiveLimit(max, penalty int) int {
	if penalty <= 0 {
		return max
	}
	divisor := math.Pow(2, float64(penalty/3))
	eff := int(float64(max) / divisor)
	if eff < 1 {
		return 1
	}
	return eff
}

// Lockout implementa o bloqueio progressivo de credenciais (A07): a partir
// de threshold falhas consecutivas numa janela de 24h, cada nova falha
// dobra o tempo de bloqueio (base, 2×base, 4×base, ...), limitado a max.
type Lockout struct {
	client    *redis.Client
	threshold int
	base      time.Duration
	max       time.Duration
}

// NewLockout cria o controle de lockout progressivo.
func NewLockout(client *redis.Client, threshold int, base, max time.Duration) *Lockout {
	return &Lockout{client: client, threshold: threshold, base: base, max: max}
}

// LockedFor retorna quanto tempo falta de bloqueio para subject (0 se livre).
func (l *Lockout) LockedFor(ctx context.Context, subject string) (time.Duration, error) {
	ttl, err := l.client.PTTL(ctx, redisx.Key("lockout", "until", subject)).Result()
	if err != nil {
		return 0, fmt.Errorf("ratelimit: lockout ttl: %w", err)
	}
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

// RegisterFailure contabiliza uma falha e, se o limiar foi atingido,
// aplica o bloqueio. Retorna a duração do bloqueio aplicado (0 se nenhum).
func (l *Lockout) RegisterFailure(ctx context.Context, subject string) (time.Duration, error) {
	countKey := redisx.Key("lockout", "fails", subject)
	pipe := l.client.TxPipeline()
	incr := pipe.Incr(ctx, countKey)
	pipe.Expire(ctx, countKey, 24*time.Hour)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("ratelimit: lockout incr: %w", err)
	}
	d := ProgressiveDelay(int(incr.Val()), l.threshold, l.base, l.max)
	if d == 0 {
		return 0, nil
	}
	if err := l.client.Set(ctx, redisx.Key("lockout", "until", subject), "1", d).Err(); err != nil {
		return 0, fmt.Errorf("ratelimit: lockout set: %w", err)
	}
	return d, nil
}

// Reset limpa falhas e bloqueio de subject (login bem-sucedido).
func (l *Lockout) Reset(ctx context.Context, subject string) error {
	return l.client.Del(ctx, redisx.Key("lockout", "fails", subject), redisx.Key("lockout", "until", subject)).Err()
}

// ProgressiveDelay calcula o bloqueio para a n-ésima falha consecutiva.
func ProgressiveDelay(failures, threshold int, base, max time.Duration) time.Duration {
	if failures < threshold {
		return 0
	}
	exp := failures - threshold
	if exp > 20 {
		return max
	}
	d := base * time.Duration(1<<exp)
	if d > max || d <= 0 {
		return max
	}
	return d
}
