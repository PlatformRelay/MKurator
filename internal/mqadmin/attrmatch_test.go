package mqadmin

import "testing"

func TestAttributeValueMatches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		key, desired, observed string
		want                   bool
	}{
		{"pub", "enabled", "ENABLED", true},
		{"maxdepth", "5000", "5000", true},
		{"maxdepth", "5000", "5000 ", true},
		{"sharecnv", "10", "10", true},
		{"descr", "a", "b", false},
		{"topstr", "retail/orders", "retail/orders", true},
	}
	for _, tc := range cases {
		if got := AttributeValueMatches(tc.key, tc.desired, tc.observed); got != tc.want {
			t.Errorf("%s %q vs %q: got %v want %v", tc.key, tc.desired, tc.observed, got, tc.want)
		}
	}
}

func TestAttributesNeedUpdate(t *testing.T) {
	t.Parallel()
	desired := map[string]string{"maxdepth": "5000", "pub": "enabled"}
	observed := map[string]string{"maxdepth": "5000", "pub": "ENABLED"}
	if AttributesNeedUpdate(desired, observed) {
		t.Fatal("expected no update")
	}
	observed["maxdepth"] = "1000"
	if !AttributesNeedUpdate(desired, observed) {
		t.Fatal("expected update on drift")
	}
}

func TestNormalizeAttrKey(t *testing.T) {
	t.Parallel()
	if got := NormalizeAttrKey("TopicStr"); got != "topstr" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeAttrKey("MaxDepth"); got != "maxdepth" {
		t.Fatalf("got %q", got)
	}
}

func TestAttributeValueMatches_NumericNormalization(t *testing.T) {
	t.Parallel()
	if !AttributeValueMatches("maxdepth", "05000", "5000") {
		t.Fatal("expected numeric equivalence")
	}
	if AttributeValueMatches("maxdepth", "bad", "5000") {
		t.Fatal("expected mismatch for non-numeric desired")
	}
}
