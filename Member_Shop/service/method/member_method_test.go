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

func TestNormalizeMemberOpenIDTreatsZeroAsEmpty(t *testing.T) {
	if got := normalizeMemberOpenID("0"); got != "" {
		t.Fatalf("zero placeholder should be normalized to empty, got %q", got)
	}
	if got := normalizeMemberOpenID(" openid-a "); got != "openid-a" {
		t.Fatalf("real openid should be trimmed and kept, got %q", got)
	}
}
