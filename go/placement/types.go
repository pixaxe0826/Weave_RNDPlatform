// Extracted from the private R&D Fabric repository and adapted to run standalone.
package placement

import "time"

// These are minimal snapshot types, not enrollment or execution authority.
type Node struct {
	ID                    string
	InstanceID            string
	Approved              bool
	EnrollmentState       string
	HealthState           string
	LastHeartbeatAt       time.Time
	DrainRequested        bool
	Compatibility         string
	SoftwareCompatibility string
	AllocatableLimits     map[string]int
	Ports                 []PortObservation
}

// AcceleratorFacts replaces private telemetry types with a normalized interface.
// The caller must supply qualified facts; no Provider implementation is included.
type AcceleratorFacts interface {
	Ready(at time.Time) bool
	EffectiveCapacity(at time.Time) (uint64, bool)
}

type Port struct {
	Protocol string
	Number   int
}

type PortObservation struct {
	ResourceID string
	Available  bool
	ObservedAt time.Time
}

type Resource struct {
	ID           string
	Kind         string
	HealthState  string
	Capacity     map[string]int
	Allocatable  map[string]int
	Capabilities []string
	GPU          AcceleratorFacts
	Port         *Port
}

type ResourceEnvelope struct {
	Version   int
	Min       int
	Preferred int
	Max       int
}

type Requirement struct {
	Kind                 string
	Quantity             int
	Envelope             *ResourceEnvelope
	Capabilities         []string
	RequiredCapabilities []string
	MinVRAMBytes         uint64
	Port                 *Port
}

type Limits struct {
	CPURate   uint32
	MemoryMiB uint64
}

type AllocationFact struct {
	RequirementIndex, Min, Preferred, Max            int
	ResourceID, Kind                                 string
	Capacity, Allocatable, ApprovedLimit, Reserved   int
	EffectiveAvailable, Selected                     int
	CapabilitiesMatch, PreferredSatisfied, VRAMMatch bool
	MinVRAMBytes, TotalVRAMBytes                     uint64
}

type Candidate struct {
	Node              Node
	Resources         []Resource
	Occupied          map[string]bool
	Reserved          map[string]int
	Exclusions        []string
	SelectedResources []string
	Quantities        map[string]int
	Allocations       []AllocationFact
	Vector            []float64
}

type Decision struct {
	Candidates     []Candidate
	SelectedNodeID string
	Reason         string
}
