package validation

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestValidateKubernetesResourceName(t *testing.T) {
	t.Parallel()
	path := field.NewPath("metadata").Child("name")

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid label", value: "app-orders", wantErr: false},
		{name: "valid subdomain", value: "app.orders", wantErr: false},
		{name: "empty allowed", value: "", wantErr: false},
		{name: "uppercase rejected", value: "APP-Orders", wantErr: true},
		{name: "underscore rejected", value: "app_orders", wantErr: true},
		{name: "leading hyphen", value: "-app", wantErr: true},
		{name: "trailing hyphen", value: "app-", wantErr: true},
		{name: "too long", value: strings.Repeat("a", 254), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs := ValidateKubernetesResourceName(path, tt.value)
			if tt.wantErr && len(errs) == 0 {
				t.Fatalf("expected error for %q", tt.value)
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Fatalf("unexpected error for %q: %v", tt.value, errs)
			}
		})
	}
}
