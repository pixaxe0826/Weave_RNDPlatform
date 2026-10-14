// Extracted from the private R&D Fabric repository and adapted to run standalone.
package placement

import (
	"reflect"
	"testing"
	"time"
)

// All quantities and labels below are synthetic, not device configuration.
func fixture() (Node, []Resource, []Requirement, time.Time) {
	now := time.Unix(123, 0).UTC()
	n := Node{ID: "alpha", InstanceID: "owner", Approved: true, EnrollmentState: "READY", HealthState: "healthy", LastHeartbeatAt: now, AllocatableLimits: map[string]int{"cpu": 10}}
	resources := []Resource{{ID: "alpha-compute", Kind: "cpu", HealthState: "healthy", Capacity: map[string]int{"quantity": 10}, Allocatable: map[string]int{"quantity": 10}, Capabilities: []string{"compute.reference"}}}
	requirements := []Requirement{{Kind: "cpu", Quantity: 3, RequiredCapabilities: []string{"compute.reference"}}}
	return n, resources, requirements, now
}

func TestEligibilityExclusions(t *testing.T) {
	for _, c := range []struct {
		name, reason string
		edit         func(*Node, []Resource, []Requirement, map[string]bool, map[string]int, time.Time)
	}{
		{"unapproved", "NODE_NOT_READY", func(n *Node, _ []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			n.Approved = false
		}},
		{"missing owner", "NODE_NOT_READY", func(n *Node, _ []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			n.InstanceID = ""
		}},
		{"stale heartbeat", "NODE_OFFLINE", func(n *Node, _ []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, now time.Time) {
			n.LastHeartbeatAt = now.Add(-6 * time.Second)
		}},
		{"future heartbeat", "NODE_OFFLINE", func(n *Node, _ []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, now time.Time) {
			n.LastHeartbeatAt = now.Add(2 * time.Second)
		}},
		{"unhealthy host", "NODE_OFFLINE", func(n *Node, _ []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			n.HealthState = "unhealthy"
		}},
		{"drain", "NODE_DRAINING", func(n *Node, _ []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			n.DrainRequested = true
		}},
		{"protocol", "PROTOCOL_INCOMPATIBLE", func(n *Node, _ []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			n.Compatibility = "incompatible"
		}},
		{"software protocol", "PROTOCOL_INCOMPATIBLE", func(n *Node, _ []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			n.SoftwareCompatibility = "incompatible"
		}},
		{"capability", "CAPABILITY_MISSING", func(_ *Node, r []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			r[0].Capabilities = nil
		}},
		{"enforcement capability", "ENFORCEMENT_UNSUPPORTED", func(_ *Node, _ []Resource, q []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			q[0].RequiredCapabilities = []string{"enforcement.reference"}
		}},
		{"capacity", "RESOURCE_LIMIT", func(_ *Node, r []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			r[0].Capacity["quantity"] = 2
		}},
		{"allocatable", "RESOURCE_LIMIT", func(_ *Node, r []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			r[0].Allocatable["quantity"] = 2
		}},
		{"approved limit", "RESOURCE_LIMIT", func(n *Node, _ []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			n.AllocatableLimits["cpu"] = 2
		}},
		{"unknown capacity", "RESOURCE_LIMIT", func(_ *Node, r []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			r[0].Capacity = nil
		}},
		{"reservation", "RESOURCE_BUSY", func(_ *Node, _ []Resource, _ []Requirement, _ map[string]bool, reserved map[string]int, _ time.Time) {
			reserved["cpu"] = 8
		}},
		{"occupied", "RESOURCE_BUSY", func(_ *Node, _ []Resource, _ []Requirement, occupied map[string]bool, _ map[string]int, _ time.Time) {
			occupied["alpha-compute"] = true
		}},
		{"unhealthy resource", "RESOURCE_UNHEALTHY", func(_ *Node, r []Resource, _ []Requirement, _ map[string]bool, _ map[string]int, _ time.Time) {
			r[0].HealthState = "unhealthy"
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			n, r, q, now := fixture()
			occupied, reserved := map[string]bool{}, map[string]int{}
			c.edit(&n, r, q, occupied, reserved, now)
			v := Evaluate(&n, r, q, occupied, reserved, nil, now, 5*time.Second)
			if !reflect.DeepEqual(v.Exclusions, []string{c.reason}) {
				t.Fatal("incorrect exclusion", v.Exclusions, c.reason)
			}
			if Rank([]Candidate{v}).SelectedNodeID != "" {
				t.Fatal("excluded candidate selected")
			}
		})
	}
}

func TestEnvelopeAndResourceOrder(t *testing.T) {
	n, resources, requirements, now := fixture()
	requirements[0].Envelope = &ResourceEnvelope{Version: 1, Min: 3, Preferred: 7, Max: 9}
	second := resources[0]
	second.ID = "beta-compute"
	for _, input := range [][]Resource{{second, resources[0]}, {resources[0], second}} {
		v := Evaluate(&n, input, requirements, nil, map[string]int{"cpu": 5}, nil, now, 5*time.Second)
		if len(v.Exclusions) != 0 || !reflect.DeepEqual(v.SelectedResources, []string{"alpha-compute"}) || v.Quantities["alpha-compute"] != 5 {
			t.Fatal("selection depends on resource order", v)
		}
		if v.Allocations[0].PreferredSatisfied {
			t.Fatal("minimum allocation mistaken for preferred allocation")
		}
	}
}

func TestStableRankingAcrossInputOrders(t *testing.T) {
	alpha := Candidate{Node: Node{ID: "alpha"}, Vector: []float64{2, 4}}
	beta := Candidate{Node: Node{ID: "beta"}, Vector: []float64{2, 4}}
	gamma := Candidate{Node: Node{ID: "gamma"}, Vector: []float64{3, 0}}
	for _, input := range [][]Candidate{
		{alpha, beta, gamma}, {alpha, gamma, beta}, {beta, alpha, gamma},
		{beta, gamma, alpha}, {gamma, alpha, beta}, {gamma, beta, alpha},
	} {
		d := Rank(input)
		if d.SelectedNodeID != "alpha" || d.Reason != "STABLE_NODE_ID" {
			t.Fatal("unstable tie-break", d)
		}
	}
	beta.Vector = []float64{1, 9}
	if d := Rank([]Candidate{alpha, beta}); d.SelectedNodeID != "beta" || d.Reason != "LEXICOGRAPHIC_COST" {
		t.Fatal("ranking ignored vector priority", d)
	}
	beta.Exclusions = []string{"RESOURCE_BUSY"}
	if d := Rank([]Candidate{beta, alpha}); d.SelectedNodeID != "alpha" || d.Reason != "ONLY_ELIGIBLE_NODE" {
		t.Fatal("excluded lower-cost candidate selected", d)
	}
	if d := Rank(nil); d.SelectedNodeID != "" || d.Reason != "NO_ELIGIBLE_NODE" {
		t.Fatal("empty input selected work", d)
	}
}

// Provider behavior is replaced only in tests; this is not GPU qualification.
type acceleratorFake struct {
	ready    bool
	capacity uint64
}

func (a acceleratorFake) Ready(time.Time) bool { return a.ready }
func (a acceleratorFake) EffectiveCapacity(time.Time) (uint64, bool) {
	return a.capacity, a.ready
}

func TestNormalizedAcceleratorFacts(t *testing.T) {
	for _, c := range []struct {
		name   string
		facts  AcceleratorFacts
		reason string
	}{
		{"known", acceleratorFake{true, 13}, ""},
		{"insufficient", acceleratorFake{true, 2}, "RESOURCE_VRAM_INSUFFICIENT"},
		{"unhealthy", acceleratorFake{false, 13}, "RESOURCE_UNHEALTHY"},
		{"unknown", nil, "RESOURCE_VRAM_INSUFFICIENT"},
	} {
		t.Run(c.name, func(t *testing.T) {
			n, r, _, now := fixture()
			n.AllocatableLimits["gpu"] = 1
			r[0].Kind, r[0].GPU = "gpu", c.facts
			q := []Requirement{{Kind: "gpu", Quantity: 1, MinVRAMBytes: 3}}
			v := Evaluate(&n, r, q, nil, nil, nil, now, 5*time.Second)
			if c.reason == "" && len(v.Exclusions) != 0 || c.reason != "" && !reflect.DeepEqual(v.Exclusions, []string{c.reason}) {
				t.Fatal(v.Exclusions)
			}
		})
	}
}

func TestLimitsAndPortFreshness(t *testing.T) {
	n, r, q, now := fixture()
	v := Evaluate(&n, r, q, nil, nil, &Limits{CPURate: 9000}, now, 5*time.Second)
	if !reflect.DeepEqual(v.Exclusions, []string{"RESOURCE_LIMIT"}) {
		t.Fatal("hard CPU limit ignored", v.Exclusions)
	}
	n.Ports = []PortObservation{{ResourceID: "alpha-service", Available: true, ObservedAt: now}}
	if !portReady(&n, "alpha-service") {
		t.Fatal("fresh observation rejected")
	}
	n.Ports[0].ObservedAt = now.Add(-4 * time.Second)
	if portReady(&n, "alpha-service") {
		t.Fatal("stale observation accepted")
	}
	if (Requirement{Quantity: 0}).Bounds().Min != 1 {
		t.Fatal("default quantity changed")
	}
}
