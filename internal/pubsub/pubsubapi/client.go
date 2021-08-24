package pubsubapi

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/broadinstitute/revere/internal/configuration"
)

// Client within the pubsub package returns the type provided by the Google
// Pub/Sub client library
func Client(config *configuration.Config) (*pubsub.Client, error) {
	client, err := pubsub.NewClient(context.Background(), config.Pubsub.ProjectID)
	return client, err
}
