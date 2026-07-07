package mqadmin

import (
	"testing"

	"pgregory.net/rapid"
)

func TestAttributeValueMatches_Reflexive(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		key := rapid.String().Draw(t, "key")
		value := rapid.String().Draw(t, "value")
		if !AttributeValueMatches(key, value, value) {
			t.Fatalf("expected reflexive match for key=%q value=%q", key, value)
		}
	})
}

func TestAttributeValueMatches_NumericSymmetric(t *testing.T) {
	t.Parallel()
	numericKeys := []string{"maxdepth", "bothresh"}
	rapid.Check(t, func(t *rapid.T) {
		key := rapid.SampledFrom(numericKeys).Draw(t, "key")
		a := rapid.String().Draw(t, "a")
		b := rapid.String().Draw(t, "b")
		ab := AttributeValueMatches(key, a, b)
		ba := AttributeValueMatches(key, b, a)
		if ab != ba {
			t.Fatalf("expected symmetry for key=%q a=%q b=%q: ab=%v ba=%v", key, a, b, ab, ba)
		}
	})
}

func TestNormalizeAttrKey_Idempotent(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		key := rapid.String().Draw(t, "key")
		once := NormalizeAttrKey(key)
		twice := NormalizeAttrKey(once)
		if once != twice {
			t.Fatalf("NormalizeAttrKey not idempotent: key=%q once=%q twice=%q", key, once, twice)
		}
	})
}

func TestAttributeValueMatches_KeyNormalizationIdempotent(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		key := rapid.String().Draw(t, "key")
		desired := rapid.String().Draw(t, "desired")
		observed := rapid.String().Draw(t, "observed")
		once := AttributeValueMatches(key, desired, observed)
		twice := AttributeValueMatches(NormalizeAttrKey(key), desired, observed)
		if once != twice {
			t.Fatalf("key normalization not idempotent in effect: key=%q desired=%q observed=%q", key, desired, observed)
		}
	})
}
