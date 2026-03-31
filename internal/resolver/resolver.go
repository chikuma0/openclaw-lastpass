package resolver

import (
	"context"
	"fmt"

	"github.com/openclaw/openclaw-lastpass/internal/config"
)

type Client interface {
	Resolve(ctx context.Context, entry string, field config.FieldSelector) (string, error)
}

type Resolver struct {
	config *config.Config
	client Client
}

func New(cfg *config.Config, client Client) *Resolver {
	return &Resolver{
		config: cfg,
		client: client,
	}
}

func (r *Resolver) Resolve(ctx context.Context, id string) (string, error) {
	mapping, ok := r.config.Lookup(id)
	if !ok {
		return "", fmt.Errorf("mapping not found for %q", id)
	}

	field, err := config.ParseField(mapping.Field)
	if err != nil {
		return "", fmt.Errorf("invalid field mapping for %q: %w", id, err)
	}

	value, err := r.client.Resolve(ctx, mapping.Entry, field)
	if err != nil {
		return "", err
	}

	return value, nil
}
