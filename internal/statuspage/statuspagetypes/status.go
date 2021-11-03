package statuspagetypes

import (
	"fmt"
)

// Status represents the various component states understood by Statuspage.io
type Status int

// https://golang.org/ref/spec#Iota, https://golang.org/ref/spec#Constant_declarations
const (
	Operational Status = iota
	DegradedPerformance
	PartialOutage
	MajorOutage
	UnderMaintenance
)

func (s Status) ToString() string {
	switch s {
	case Operational:
		return "Operational"
	case DegradedPerformance:
		return "Degraded Performance"
	case PartialOutage:
		return "Partial Outage"
	case MajorOutage:
		return "Major Outage"
	case UnderMaintenance:
		return "Under Maintenance"
	}
	return fmt.Sprintf("Invalid Status %d", s)
}

func (s Status) ToSnakeCase() string {
	switch s {
	case Operational:
		return "operational"
	case DegradedPerformance:
		return "degraded_performance"
	case PartialOutage:
		return "partial_outage"
	case MajorOutage:
		return "major_outage"
	case UnderMaintenance:
		return "under_maintenance"
	}
	return fmt.Sprintf("invalid_status_%d", s)
}

func StatusFromKebabCase(kebabCaseString string) (Status, error) {
	switch kebabCaseString {
	case "operational":
		return Operational, nil
	case "degraded-performance":
		return DegradedPerformance, nil
	case "partial-outage":
		return PartialOutage, nil
	case "major-outage":
		return MajorOutage, nil
	case "under-maintenance":
		return UnderMaintenance, nil
	}
	return -1, fmt.Errorf("%s cannot be parsed to a Status", kebabCaseString)
}

// WorstWith (ab)uses the fact that iota increments per field, so a higher int value
// is further down the list
func (s Status) WorstWith(other Status) Status {
	if s < other {
		return other
	} else {
		return s
	}
}
