// Extracted from the private R&D Fabric repository and adapted to run standalone.
package continuity

import (
	"math"
	"testing"
)

func TestCheckedMonotonicGeneration(t *testing.T) {
	for _, c := range []struct{ last, floor, want uint64 }{{8, 26, 27}, {28, 26, 29}, {0, 0, 1}} {
		got, e := NextGeneration(c.last, c.floor)
		if e != nil || got != c.want {
			t.Fatal(c, got, e)
		}
	}
	for _, c := range [][2]uint64{{math.MaxUint64, 0}, {0, math.MaxUint64}} {
		if _, e := NextGeneration(c[0], c[1]); e == nil {
			t.Fatal("overflow accepted")
		}
	}
}

// The fake models snapshot/commit boundaries only, not production durability.
type memoryStore struct{ state State }

func (m *memoryStore) update(next State) error {
	if err := ValidateTransition(m.state, next); err != nil {
		return err
	}
	m.state = next
	return nil
}

func TestMonotonicGuardAndRestartSnapshot(t *testing.T) {
	initial := State{ControllerID: "alpha", Generation: 4, ContinuityFloor: 9}
	for _, c := range []struct {
		name string
		next State
	}{
		{"lower floor", State{"alpha", 4, 8}},
		{"lower generation", State{"alpha", 3, 9}},
		{"different authority", State{"beta", 4, 9}},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := memoryStore{initial}
			if m.update(c.next) == nil || m.state != initial {
				t.Fatal("rejected transition changed snapshot")
			}
		})
	}
	store := memoryStore{initial}
	next := initial
	next.ContinuityFloor = 12
	if err := store.update(next); err != nil {
		t.Fatal(err)
	}
	// Recreate the test fake from the retained committed snapshot.
	restarted := memoryStore{store.state}
	if restarted.state.ContinuityFloor != 12 {
		t.Fatal("restart lost floor")
	}
	if err := restarted.update(initial); err == nil {
		t.Fatal("restart permitted rollback")
	}
	generation, err := NextGeneration(restarted.state.Generation, restarted.state.ContinuityFloor)
	if err != nil || generation != 13 {
		t.Fatal("retained floor did not constrain allocation", generation, err)
	}
	if err := restarted.update(restarted.state); err != nil {
		t.Fatal("unchanged state rejected", err)
	}
}

func TestHighWaterUnknownIsNotZero(t *testing.T) {
	for _, c := range []struct {
		name, input string
	}{
		{"missing", `{}`},
		{"null map", `{"high_water":null}`},
		{"null value", `{"high_water":{"alpha":null}}`},
		{"negative", `{"high_water":{"alpha":-1}}`},
		{"fraction", `{"high_water":{"alpha":1.5}}`},
		{"wrong type", `{"high_water":{"alpha":"unknown"}}`},
		{"malformed", `{"high_water":`},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := ReadHighWater([]byte(c.input), "high_water"); err == nil {
				t.Fatal("unsafe or unknown high-water accepted")
			}
		})
	}
	for _, input := range []string{`{"high_water":{}}`, `{"high_water":{"alpha":0}}`} {
		if _, err := ReadHighWater([]byte(input), "high_water"); err != nil {
			t.Fatal("explicit known zero rejected", err)
		}
	}
}
