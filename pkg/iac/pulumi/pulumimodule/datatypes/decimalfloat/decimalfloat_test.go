package decimalfloat

import "testing"

func TestFloat64KeepsTheDecimal(t *testing.T) {
	for _, c := range []struct {
		in   float32
		want float64
	}{
		{-0.3, -0.3},
		{0.5, 0.5},
		{0.1, 0.1},
		{33.3, 33.3},
		{1, 1},
		{0, 0},
	} {
		if got := Float64(c.in); got != c.want {
			t.Errorf("Float64(%v) = %v, want %v", c.in, got, c.want)
		}
	}
	if float64(float32(-0.3)) == -0.3 {
		t.Fatal("the plain widening no longer drifts; this package may be unnecessary")
	}
}
