package pubsubtypes

import "github.com/broadinstitute/revere/internal/cloudmonitoring"

// PerComponentHandler is an alias for a function handling the update of a single status.
// It is abstracted so the type may be referenced in across the program without importing other code.
type PerComponentHandler func(componentName string, labels *cloudmonitoring.AlertLabels, incident *cloudmonitoring.MonitoringIncident) error
