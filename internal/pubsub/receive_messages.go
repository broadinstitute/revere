package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"
)

// receiveOnce should handle and acknowledge a single message; will run
// asynchronously
func receiveOnce(config *configuration.Config, msg *pubsub.Message) {
	fmt.Printf("Message %q from %s", string(msg.Data), config.Pubsub.SubscriptionID)
	msg.Ack()
}

// ReceiveMessages should never terminate, it continually pulls messages from
// the subscription
func ReceiveMessages(config *configuration.Config, client *pubsub.Client, ctx context.Context) error {
	subscription := client.Subscription(config.Pubsub.SubscriptionID)
	return subscription.Receive(ctx, func(cctx context.Context, msg *pubsub.Message) {
		receiveOnce(config, msg)
	})
}
