package seeding

import "testing"

func TestAutoScalerRecommendBots(t *testing.T) {
	s := NewAutoScaler(1, 6)

	cases := []struct {
		users int
		want  int
	}{
		{0, 5},
		{25, 6},
		{120, 4},
		{350, 3},
		{900, 1},
	}

	for _, tc := range cases {
		got := s.RecommendBots(tc.users)
		if got != tc.want {
			t.Fatalf("users=%d: got %d want %d", tc.users, got, tc.want)
		}
	}
}
