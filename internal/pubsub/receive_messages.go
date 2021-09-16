package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/revere/internal/cloudmonitoring"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/shared"
)

// receiveOnce should handle and acknowledge a single message; will run
// asynchronously
func receiveOnce(config *configuration.Config, msg *pubsub.Message) {
	var packet *cloudmonitoring.MonitoringPacket
	if err := json.Unmarshal(msg.Data, &packet); err != nil {
		shared.LogLn(config, "failed to parse packet", fmt.Sprintf("%+v", err))
		return
	}
	labels, err := packet.ParseLabels()
	if err != nil {
		shared.LogLn(config, "failed to parse labels", fmt.Sprintf("%+v", err))
		return
	}
	fmt.Printf("Alert %+v from %s", labels, config.Pubsub.SubscriptionID)
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
