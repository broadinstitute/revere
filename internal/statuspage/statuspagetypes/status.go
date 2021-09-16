package statuspagetypes

import "fmt"

type Status int

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
