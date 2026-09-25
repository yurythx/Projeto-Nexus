package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yurythx/projeto-nexus/internal/modules/signum/domain"
	"github.com/yurythx/projeto-nexus/internal/platform/passwords"
)

// KeycloakCredentials devolve issuer, client_id e client_secret do client
// usado na reautenticação (lido da configuração dinâmica do Keycloak com
// fallback para o .env).
type KeycloakCredentials func(ctx context.Context) (issuer, clientID, clientSecret string)

// Reauthenticator confere a senha no momento da assinatura:
//   - conta local: Argon2id/bcrypt contra users.password_hash;
//   - conta federada: grant "password" (Direct Access Grants) no Keycloak
//     dedicado — que por sua vez valida no Active Directory via LDAP(S).
type Reauthenticator struct {
	pool   *pgxpool.Pool
	creds  KeycloakCredentials
	client *http.Client
}

// NewReauthenticator cria o reautenticador.
func NewReauthenticator(pool *pgxpool.Pool, creds KeycloakCredentials) *Reauthenticator {
	return &Reauthenticator{pool: pool, creds: creds, client: &http.Client{Timeout: 10 * time.Second}}
}

var _ domain.Reauthenticator = (*Reauthenticator)(nil)

// Reauthenticate implementa domain.Reauthenticator.
func (r *Reauthenticator) Reauthenticate(ctx context.Context, userID uuid.UUID, username string, federated bool, password string) (string, error) {
	if password == "" {
		return "", domain.ErrReauth
	}
	if !federated {
		var hash *string
		if err := r.pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1 AND active`, userID).Scan(&hash); err != nil || hash == nil {
			passwords.DummyVerify(password)
			return "", domain.ErrReauth
		}
		if err := passwords.Verify(*hash, password); err != nil {
			return "", domain.ErrReauth
		}
		return "password_reauth:local", nil
	}

	issuer, clientID, secret := r.creds(ctx)
	if issuer == "" || clientID == "" {
		return "", fmt.Errorf("signum: Keycloak não configurado para reautenticação")
	}
	form := url.Values{
		"grant_type": {"password"},
		"client_id":  {clientID},
		"username":   {username},
		"password":   {password},
		"scope":      {"openid"},
	}
	if secret != "" {
		form.Set("client_secret", secret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimSuffix(issuer, "/")+"/protocol/openid-connect/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("signum: Keycloak indisponível: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
	switch {
	case resp.StatusCode == http.StatusOK:
		return "password_reauth:keycloak", nil
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest:
		return "", domain.ErrReauth
	default:
		return "", errors.New("signum: resposta inesperada do Keycloak na reautenticação")
	}
}
