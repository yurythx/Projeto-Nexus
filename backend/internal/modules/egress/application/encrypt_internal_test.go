package application

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	apperrors "github.com/yurythx/projeto-nexus/internal/domain/errors"
	"github.com/yurythx/projeto-nexus/internal/modules/egress/domain"
	"github.com/yurythx/projeto-nexus/internal/platform/netguard"
)

type openPolicy struct{}

func (openPolicy) Deliver(context.Context, domain.Target, domain.Delivery) (int, error) {
	return 0, nil
}
func (openPolicy) Policy() netguard.Policy {
	return netguard.Policy{AllowPrivate: true, AllowHTTP: true}
}

// Falha ao cifrar o segredo é erro interno, e nada é gravado.
func TestSecretEncryptionFailureIsInternal(t *testing.T) {
	s := &Service{deliverer: openPolicy{}, encrypt: func(string) (string, error) { return "", errors.New("sem entropia") }}
	secret := "x"
	_, err := s.SaveTarget(context.Background(), uuid.Nil, TargetInput{Name: "x", Kind: "webhook", URL: "http://127.0.0.1/x", Secret: &secret})
	if appErr, ok := apperrors.As(err); !ok || appErr.Status != http.StatusInternalServerError {
		t.Fatalf("falha de cifragem: %v", err)
	}
}
