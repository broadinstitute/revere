package cloudmonitoring

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
)

type AlertLabels struct {
	ServiceName        string
	ServiceEnvironment string
	AlertType          statuspagetypes.Status
}

func (p *MonitoringPacket) ParseLabels() (*AlertLabels, error) {
	serviceName, present := p.Incident.PolicyUserLabels["revere-service-name"]
	if !present {
		return nil, fmt.Errorf("alert labels lacked the service name in %+v", p)
	}
	serviceEnvironment, present := p.Incident.PolicyUserLabels["revere-service-environment"]
	if !present {
		return nil, fmt.Errorf("alert labels lacked the service environment in %+v", p)
	}
	alertTypeString, present := p.Incident.PolicyUserLabels["revere-alert-type"]
	if !present {
		return nil, fmt.Errorf("alert labels lacked the alert type in %+v", p)
	}
	alertType, err := statuspagetypes.StatusFromKebabCase(alertTypeString)
	if err != nil {
		return nil, fmt.Errorf("alert label's alert type incorrect format: %w", err)
	}
	return &AlertLabels{
		ServiceName:        serviceName,
		ServiceEnvironment: serviceEnvironment,
		AlertType:          alertType,
	}, nil
}
