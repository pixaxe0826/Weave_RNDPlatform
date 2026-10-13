// Extracted from the private R&D Fabric repository and adapted to run standalone.
package continuity

import (
	"encoding/json"
	"errors"
	"math"
)

// State contains only the fields used by the monotonic transition guard.
// Persistence and identity-transfer authorization are outside this sample.
// Inputs must already be validated snapshots.
type State struct {
	ControllerID    string
	Generation      uint64
	ContinuityFloor uint64
}

func NextGeneration(last, floor uint64) (uint64, error) {
	n := max(last, floor)
	if n == math.MaxUint64 {
		return 0, errors.New("LEASE_GENERATION_EXHAUSTED")
	}
	return n + 1, nil
}

// ValidateTransition is the pure guard, without the surrounding database update.
func ValidateTransition(prior, state State) error {
	if prior.ControllerID != state.ControllerID || state.ContinuityFloor < prior.ContinuityFloor || state.Generation < prior.Generation {
		return errors.New("CONTINUITY_MONOTONIC_STATE_VIOLATION")
	}
	return nil
}

// JSON null must not become uint64(0). The producer persists integer fences.
func ReadHighWater(b []byte, field string) (map[string]uint64, error) {
	var root map[string]json.RawMessage
	if json.Unmarshal(b, &root) != nil {
		return nil, errors.New("CONTINUITY_JOURNAL_FORMAT")
	}
	var raw map[string]json.RawMessage
	if json.Unmarshal(root[field], &raw) != nil || raw == nil {
		return nil, errors.New("CONTINUITY_HIGH_WATER_UNKNOWN")
	}
	out := map[string]uint64{}
	for id, value := range raw {
		var n uint64
		if string(value) == "null" || json.Unmarshal(value, &n) != nil {
			return nil, errors.New("CONTINUITY_HIGH_WATER_UNKNOWN")
		}
		out[id] = n
	}
	return out, nil
}
