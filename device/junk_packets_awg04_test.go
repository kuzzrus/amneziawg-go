package device

import "testing"

// TestJunkPackets_MinExceedsMax_AWG04 is a regression test for the
// unsigned-underflow bug: max-min wrapped to near 2^32 when min > max,
// making make([]byte, min+fastrandn(max-min)) attempt a multi-gigabyte
// allocation. JunkPackets must now clamp instead of underflow.
func TestJunkPackets_MinExceedsMax_AWG04(t *testing.T) {
	cases := []struct {
		name     string
		min, max uint32
		count    uint32
	}{
		{"min > max (the exact AWG-04 bug shape)", 100, 40, 5},
		{"min set, max left at its zero-value default", 100, 0, 5},
		{"min == max (must stay allowed, not rejected)", 50, 50, 5},
		{"min < max (ordinary case)", 40, 100, 5},
		{"min == max == 0 (junk disabled)", 0, 0, 5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := &Device{}
			d.junk.min.Store(c.min)
			d.junk.max.Store(c.max)
			d.junk.count.Store(c.count)

			bufs := d.JunkPackets()

			if len(bufs) != int(c.count) {
				t.Fatalf("got %d packets, want %d", len(bufs), c.count)
			}
			lo := c.min
			hi := c.max
			if hi < lo {
				hi = lo
			}
			for i, b := range bufs {
				if uint32(len(b)) < lo || uint32(len(b)) > hi {
					t.Errorf("packet %d: len=%d, want in [%d,%d] -- min>max must clamp, not underflow into a huge allocation", i, len(b), lo, hi)
				}
			}
		})
	}
}
