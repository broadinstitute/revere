package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/revere/internal/cloudmonitoring"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/pubsub/pubsubtypes"
	"github.com/broadinstitute/revere/internal/shared"
	"os"
)

// receiveOnce should handle a single message; will run asynchronously
func receiveOnce(config *configuration.Config, msg *pubsub.Message, callback pubsubtypes.PerComponentHandler) error {
	// parse Google's data structure
	var packet *cloudmonitoring.MonitoringPacket
	if err := json.Unmarshal(msg.Data, &packet); err != nil {
		shared.LogLn(config, fmt.Sprintf("failed to parse packet, ignoring: %v", err))
		return nil
	}

	// parse Revere's labels
	labels, err := packet.ParseLabels()
	if err != nil {
		shared.LogLn(config, fmt.Sprintf("failed to parse labels from %s packet, ignoring: %v", packet.Incident.PolicyName, err))
		return nil
	}
	shared.LogLn(config, fmt.Sprintf("pubsub alert %s (closed: %v) -- parsed %+v (%s)",
		packet.Incident.PolicyName, packet.Incident.HasEnded(), labels, labels.AlertType.ToString()))

	// execute callback for each affected component
	affectedSomeComponents := false
	for _, serviceMapping := range config.ServiceToComponentMapping {
		if serviceMapping.ServiceName == labels.ServiceName &&
			serviceMapping.ServiceEnvironment == labels.ServiceEnvironment {
			for _, componentName := range serviceMapping.AffectsComponentsNamed {
				affectedSomeComponents = true
				shared.LogLn(config,
					fmt.Sprintf("pubsub alert %s affects %s, executing callback...", packet.Incident.IncidentID, componentName))
				if err := callback(componentName, labels, packet.Incident); err != nil {
					shared.LogLn(config,
						"failed to execute callback", fmt.Sprintf("%+v", err))
					return err
				}
			}
		}
	}
	if !affectedSomeComponents {
		shared.LogLn(config, fmt.Sprintf("pubsub alert %s affected no components, ignoring", packet.Incident.PolicyName))
	}
	return nil
}

// ReceiveMessages should never terminate, it continually pulls messages from the subscription
func ReceiveMessages(config *configuration.Config, client *pubsub.Client, ctx context.Context, callback pubsubtypes.PerComponentHandler) error {
	subscription := client.Subscription(config.Pubsub.SubscriptionID)
	return subscription.Receive(ctx, func(cctx context.Context, msg *pubsub.Message) {
		if err := receiveOnce(config, msg, callback); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			// There may be multiple infinite goroutines (see serve.go), to exit we have to do so forcibly.
			// We specifically don't call msg.Nack before doing so because we want to let the lease expire,
			// instead of pubsub retrying it immediately and *then* having it time out because we've exited.
			os.Exit(1)
		} else {
			msg.Ack()
		}
	})
}
