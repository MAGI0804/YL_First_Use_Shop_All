package method

import "testing"

func TestUniqueUintRemovesZeroAndDuplicates(t *testing.T) {
	got := uniqueUint([]uint{2, 0, 2, 5, 5})
	if len(got) != 2 || got[0] != 2 || got[1] != 5 {
		t.Fatalf("uniqueUint returned %v", got)
	}
}

func TestNormalizeMemberDefaults(t *testing.T) {
	if normalizeMemberStatus("") != "active" {
		t.Fatalf("empty status should default active")
	}
	if normalizeMemberSource("") != "backend" {
		t.Fatalf("empty source should default backend")
	}
}
