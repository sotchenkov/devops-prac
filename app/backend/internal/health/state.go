package health

import "sync"

type Phase string

const (
	PhaseStarting    Phase = "starting"
	PhaseReady       Phase = "ready"
	PhaseTerminating Phase = "terminating"
)

type State struct {
	mu    sync.RWMutex
	phase Phase
}

func New() *State {
	return &State{phase: PhaseStarting}
}

func (s *State) MarkReady() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.phase == PhaseTerminating {
		return false
	}
	s.phase = PhaseReady
	return true
}

func (s *State) BeginTermination() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phase = PhaseTerminating
}

func (s *State) Snapshot() (bool, Phase) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.phase == PhaseReady, s.phase
}
