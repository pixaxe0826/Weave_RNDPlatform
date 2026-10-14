// Extracted from the private R&D Fabric repository and adapted to run standalone.
package placement

import (
	"sort"
	"strings"
	"time"
)

// Evaluate extracts Node/resource eligibility from an already normalized snapshot.
// Storage/network eligibility, score-vector construction and Lease issuance are
// outside this sample. The caller supplies the policy timeout and reservations.
// It explains exclusions without granting Jobs or changing admission state.
func Evaluate(n *Node, resources []Resource, requirements []Requirement, occupied map[string]bool, reserved map[string]int, limits *Limits, now time.Time, heartbeatTimeout time.Duration) Candidate {
	v := Candidate{Node: *n, Resources: append([]Resource{}, resources...), Occupied: occupied, Reserved: reserved, Exclusions: []string{}, SelectedResources: []string{}, Quantities: map[string]int{}}
	sort.Slice(v.Resources, func(i, k int) bool { return v.Resources[i].ID < v.Resources[k].ID })
	add := func(code string) {
		if !contains(v.Exclusions, code) {
			v.Exclusions = append(v.Exclusions, code)
		}
	}
	if n.DrainRequested || n.EnrollmentState == "DRAINING" {
		add("NODE_DRAINING")
	}
	if n.EnrollmentState == "OFFLINE" || n.HealthState != "healthy" || now.Sub(n.LastHeartbeatAt) > heartbeatTimeout || n.LastHeartbeatAt.After(now.Add(time.Second)) {
		add("NODE_OFFLINE")
	}
	if !n.Approved || n.EnrollmentState != "READY" || n.InstanceID == "" {
		add("NODE_NOT_READY")
	}
	if n.Compatibility == "incompatible" || n.SoftwareCompatibility == "incompatible" {
		add("PROTOCOL_INCOMPATIBLE")
	}
	for qi, q := range requirements {
		caps := append(append([]string{}, q.Capabilities...), q.RequiredCapabilities...)
		fit, vram, capacity, healthy, free, enforced := false, false, false, false, false, false
		for _, r := range v.Resources {
			if r.Kind != q.Kind {
				continue
			}
			fact := allocationFact(q, r, n, v.Reserved[q.Kind], qi)
			v.Allocations = append(v.Allocations, fact)
			if enforced {
				continue
			}
			if !fact.CapabilitiesMatch {
				continue
			}
			fit = true
			if !fact.VRAMMatch {
				continue
			}
			vram = true
			quantity := fact.Selected
			if quantity < fact.Min {
				continue
			}
			capacity = true
			if r.HealthState != "healthy" || r.GPU != nil && !r.GPU.Ready(now) {
				continue
			}
			healthy = true
			if occupied[r.ID] || q.Kind == "port" && (q.Port == nil || r.Port == nil || q.Port.Protocol != r.Port.Protocol || q.Port.Number != 0 && q.Port.Number != r.Port.Number || !portReady(n, r.ID)) {
				continue
			}
			free = true
			// CPU rates use basis points: 10000 represents 100% of resource capacity.
			if limits != nil && (q.Kind == "cpu" && uint64(limits.CPURate)*uint64(r.Capacity["quantity"]) > uint64(quantity)*10000 || q.Kind == "memory" && limits.MemoryMiB > uint64(quantity)) {
				continue
			}
			enforced = true
			v.SelectedResources = append(v.SelectedResources, r.ID)
			v.Quantities[r.ID] = quantity
		}
		switch {
		case !fit:
			code := "CAPABILITY_MISSING"
			for _, cap := range caps {
				if strings.HasPrefix(cap, "enforcement.") {
					code = "ENFORCEMENT_UNSUPPORTED"
				}
			}
			add(code)
		case !vram:
			add("RESOURCE_VRAM_INSUFFICIENT")
		case !capacity:
			if v.Reserved[q.Kind] > 0 {
				add("RESOURCE_BUSY")
			} else {
				add("RESOURCE_LIMIT")
			}
		case !healthy:
			add("RESOURCE_UNHEALTHY")
		case !free:
			add("RESOURCE_BUSY")
		case !enforced:
			add("RESOURCE_LIMIT")
		}
	}
	sort.Strings(v.Exclusions)
	return v
}

// Rank compares caller-supplied, equal-length finite score vectors. It does not
// build cache/network/telemetry scores or allocate resources.
// All vectors must have the same length and contain only finite values (no NaN or infinity).
func Rank(candidates []Candidate) Decision {
	d := Decision{Candidates: candidates, Reason: "NO_ELIGIBLE_NODE"}
	choices := []Candidate{}
	for _, v := range d.Candidates {
		if len(v.Exclusions) == 0 {
			choices = append(choices, v)
		}
	}
	sort.Slice(choices, func(i, k int) bool { return rankLess(choices[i], choices[k]) })
	if len(choices) > 0 {
		d.SelectedNodeID = choices[0].Node.ID
		d.Reason = "ONLY_ELIGIBLE_NODE"
		if len(choices) > 1 {
			d.Reason = "LEXICOGRAPHIC_COST"
			same := true
			for i, x := range choices[0].Vector {
				if x != choices[1].Vector[i] {
					same = false
				}
			}
			if same {
				d.Reason = "STABLE_NODE_ID"
			}
		}
	}
	return d
}

func rankLess(a, b Candidate) bool {
	for i, v := range a.Vector {
		if v != b.Vector[i] {
			return v < b.Vector[i]
		}
	}
	return a.Node.ID < b.Node.ID
}

func allocationFact(q Requirement, r Resource, n *Node, reserved, index int) AllocationFact {
	e := q.Bounds()
	v := AllocationFact{RequirementIndex: index, ResourceID: r.ID, Kind: q.Kind, Min: e.Min, Preferred: e.Preferred, Max: e.Max, Capacity: r.Capacity["quantity"], Allocatable: r.Allocatable["quantity"], ApprovedLimit: n.AllocatableLimits[q.Kind], Reserved: reserved, CapabilitiesMatch: true, MinVRAMBytes: q.MinVRAMBytes, VRAMMatch: true}
	v.EffectiveAvailable = max(0, min(v.Capacity, v.Allocatable, v.ApprovedLimit)-reserved)
	if v.EffectiveAvailable >= e.Min {
		v.Selected = min(e.Preferred, e.Max, v.EffectiveAvailable)
	}
	v.PreferredSatisfied = v.Selected >= e.Preferred
	for _, cap := range append(append([]string{}, q.Capabilities...), q.RequiredCapabilities...) {
		if !contains(r.Capabilities, cap) {
			v.CapabilitiesMatch = false
		}
	}
	if r.GPU != nil {
		v.TotalVRAMBytes, _ = r.GPU.EffectiveCapacity(n.LastHeartbeatAt)
	}
	v.VRAMMatch = q.MinVRAMBytes == 0 || r.GPU != nil && v.TotalVRAMBytes >= q.MinVRAMBytes
	return v
}

func (r Requirement) Bounds() ResourceEnvelope {
	if r.Envelope != nil {
		return *r.Envelope
	}
	n := r.Quantity
	if n == 0 {
		n = 1
	}
	return ResourceEnvelope{Version: 1, Min: n, Preferred: n, Max: n}
}

func portReady(n *Node, id string) bool {
	for _, p := range n.Ports {
		// Port observations expire at three seconds of age relative to the last heartbeat.
		if p.ResourceID == id && p.Available && n.LastHeartbeatAt.Sub(p.ObservedAt) >= -time.Second && n.LastHeartbeatAt.Sub(p.ObservedAt) < 3*time.Second {
			return true
		}
	}
	return false
}

func contains(v []string, s string) bool {
	for _, x := range v {
		if x == s {
			return true
		}
	}
	return false
}
