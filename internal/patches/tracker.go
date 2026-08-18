package patches

import "sync"

// Tracker remembers the most recent report per target so the control panel can
// display patch status long after the response was served.
type Tracker struct {
	mu      sync.Mutex
	reports map[Target]Report
}

// NewTracker returns an empty Tracker.
func NewTracker() *Tracker {
	return &Tracker{reports: map[Target]Report{}}
}

// Record stores the report for one target.
func (t *Tracker) Record(target Target, report Report) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reports[target] = report
}

// Results merges the recorded reports into registry order, one entry per patch.
func (t *Tracker) Results() Report {
	t.mu.Lock()
	defer t.mu.Unlock()

	seen := map[string]bool{}
	var out Report

	for _, p := range All() {
		for _, report := range t.reports {
			for _, res := range report {
				if res.ID == p.ID && !seen[res.ID] {
					seen[res.ID] = true
					out = append(out, res)
				}
			}
		}
	}
	return out
}
