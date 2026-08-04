package health

import "testing"

func TestLifecycleTransitions(t *testing.T) {
	state := New()

	ready, phase := state.Snapshot()
	if ready || phase != PhaseStarting {
		t.Fatalf("initial state = ready:%v phase:%q, want false/%q", ready, phase, PhaseStarting)
	}

	if changed := state.MarkReady(); !changed {
		t.Fatal("MarkReady() = false, want true")
	}
	ready, phase = state.Snapshot()
	if !ready || phase != PhaseReady {
		t.Fatalf("ready state = ready:%v phase:%q, want true/%q", ready, phase, PhaseReady)
	}

	state.BeginTermination()
	ready, phase = state.Snapshot()
	if ready || phase != PhaseTerminating {
		t.Fatalf("terminating state = ready:%v phase:%q, want false/%q", ready, phase, PhaseTerminating)
	}

	if changed := state.MarkReady(); changed {
		t.Fatal("MarkReady() after termination = true, want false")
	}
	ready, phase = state.Snapshot()
	if ready || phase != PhaseTerminating {
		t.Fatalf("final state = ready:%v phase:%q, want false/%q", ready, phase, PhaseTerminating)
	}
}
