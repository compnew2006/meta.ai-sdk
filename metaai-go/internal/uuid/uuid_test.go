package uuid

import (
	"strings"
	"testing"
)

func TestV4Format(t *testing.T) {
	for i := 0; i < 16; i++ {
		id := V4()
		if len(id) != 36 {
			t.Fatalf("V4() = %q, want 36 chars", id)
		}
		// version nibble at position 14 must be '4', variant at position 19 in {8,9,a,b}
		if id[14] != '4' {
			t.Errorf("V4() = %q, version nibble %q want '4'", id, string(id[14]))
		}
		switch id[19] {
		case '8', '9', 'a', 'b':
		default:
			t.Errorf("V4() = %q, variant nibble %q want one of 8,9,a,b", id, string(id[19]))
		}
	}
}

func TestV4Uniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		id := V4()
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate UUID generated: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestNumericString(t *testing.T) {
	for _, n := range []int{1, 5, 19, 2} {
		s := NumericString(n)
		if len(s) != n {
			t.Errorf("NumericString(%d) = %q (len %d), want len %d", n, s, len(s), n)
		}
		for _, r := range s {
			if r < '0' || r > '9' {
				t.Errorf("NumericString(%d) = %q, non-digit rune", n, s)
			}
		}
	}
	seen := make(map[string]struct{}, 500)
	for i := 0; i < 500; i++ {
		s := NumericString(12)
		if _, dup := seen[s]; dup {
			t.Fatalf("duplicate numeric id: %s", s)
		}
		seen[s] = struct{}{}
	}
}

func TestOfflineThreadingID(t *testing.T) {
	id := OfflineThreadingID()
	if id == "" || !strings.ContainsAny(id, "0123456789") {
		t.Errorf("OfflineThreadingID() = %q, want decimal string", id)
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			t.Errorf("OfflineThreadingID() = %q, non-digit rune %q", id, r)
		}
	}
	// Two calls close in time still differ in the random low bits.
	a, b := OfflineThreadingID(), OfflineThreadingID()
	if a == b {
		t.Errorf("OfflineThreadingID() produced identical values %s twice", a)
	}
}
