package cloudmonitoring

import "google.golang.org/genproto/googleapis/monitoring/v3"

// MonitoringPacket handles payloads from Webhook *or Pub/Sub*
// https://cloud.google.com/monitoring/support/notification-options#webhooks
// Pointers used for nested structs to avoid copying mutexes
type MonitoringPacket struct {
	Version  string              `json:"version"`
	Incident *MonitoringIncident `json:"incident"`
}

// MonitoringIncident is the union of the v1.1 and v1.2 types, because v1.2 is
// already a perfect superset of v1.1 save for the Errors field
type MonitoringIncident struct {
	// Incident
	IncidentID string `json:"incident_id"`
	URL        string `json:"url"`
	State      string `json:"state"`
	StartedAt  int64  `json:"started_at"`
	EndedAt    int64  `json:"ended_at"`
	Summary    string `json:"summary"`
	ApigeeURL  string `json:"apigee_url"`

	// Resource
	Resource                *MonitoringResource `json:"resource"`
	ResourceTypeDisplayName string              `json:"resource_type_display_name"`
	ResourceID              string              `json:"resource_id"`
	ResourceDisplayName     string              `json:"resource_display_name"`
	ResourceName            string              `json:"resource_name"`

	// Metric
	Metric *MonitoringMetric `json:"metric"`

	// Policy
	PolicyName       string                                `json:"policy_name"`
	PolicyUserLabels map[string]string                     `json:"policy_user_labels"`
	Documentation    *monitoring.AlertPolicy_Documentation `json:"documentation"`
	Condition        *monitoring.AlertPolicy_Condition     `json:"condition"`
	ConditionName    string                                `json:"condition_name"`

	// Errors (on Google's end in formulating the incident)
	// Technically google.rpc.Status objects but pulling in all of gRPC
	// to fully type errors seems pointless, we'd just dump them anyway
	Errors []map[string]interface{} `json:"errors"`
}

type MonitoringResource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
}

type MonitoringMetric struct {
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
}

func (i *MonitoringIncident) HasEnded() bool {
	switch i.State {
	case "open":
		return false
	case "closed":
		return true
	default:
		// The incident state is theoretically an enum, but failing that we return based
		// on the incident having a sensible ending time.
		// Google silently removed docs on an older version of MonitoringIncident that
		// lacked many of the fields of the modern one.
		return i.EndedAt > i.StartedAt
	}
}
