package mqadmin

import "testing"

func TestNormalizeChannelType(t *testing.T) {
	t.Parallel()
	if got := NormalizeChannelType(""); got != ChannelTypeSvrconn {
		t.Fatalf("empty = %q, want svrconn", got)
	}
	if got := NormalizeChannelType(ChannelTypeSdr); got != ChannelTypeSdr {
		t.Fatalf("sdr = %q", got)
	}
	if got := NormalizeChannelType(ChannelTypeRcvr); got != ChannelTypeRcvr {
		t.Fatalf("rcvr = %q", got)
	}
}

func TestChannelTypeSupported(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in   ChannelType
		want bool
	}{
		{"", true},
		{ChannelTypeSvrconn, true},
		{ChannelTypeSdr, true},
		{ChannelTypeRcvr, true},
		{ChannelType("clusrcv"), false},
	} {
		if got := ChannelTypeSupported(tc.in); got != tc.want {
			t.Fatalf("ChannelTypeSupported(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
