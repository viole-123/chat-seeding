package seeding

// AutoScaler recommends bot activity level based on active-user saturation.
// Goal: warm up empty rooms, amplify mid-traffic, and back off in high traffic.
type AutoScaler struct {
	MinBots int
	MaxBots int
}

func NewAutoScaler(minBots, maxBots int) *AutoScaler {
	if minBots < 1 {
		minBots = 1
	}
	if maxBots < minBots {
		maxBots = minBots
	}
	return &AutoScaler{MinBots: minBots, MaxBots: maxBots}
}

// RecommendBots returns the target number of bot messages per minute window.
func (a *AutoScaler) RecommendBots(activeUsers int) int {
	if activeUsers <= 0 {
		return clamp(a.MaxBots-1, a.MinBots, a.MaxBots)
	}

	// 0-50 users: avoid cold start by being active.
	if activeUsers <= 50 {
		return clamp(a.MaxBots, a.MinBots, a.MaxBots)
	}

	// 51-200 users: keep high but slightly reduced bot pressure.
	if activeUsers <= 200 {
		return clamp((a.MaxBots*8)/10, a.MinBots, a.MaxBots)
	}

	// 201-500 users: taper down.
	if activeUsers <= 500 {
		return clamp((a.MaxBots*5)/10, a.MinBots, a.MaxBots)
	}

	// 500+ users: let real users dominate.
	return a.MinBots
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
