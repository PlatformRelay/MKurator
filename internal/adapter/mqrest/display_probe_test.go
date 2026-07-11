package mqrest

import (
	"testing"

	"github.com/platformrelay/mkurator/internal/mqadmin"
)

func TestQueueAttributeDisplayProbeRequest(t *testing.T) {
	t.Parallel()
	req := queueAttributeDisplayProbeRequest("APP.PROBE", attrShare)
	if req.Command != "display" || req.Qualifier != "qlocal" || req.Name != "APP.PROBE" {
		t.Fatalf("request = %+v", req)
	}
	if len(req.ResponseParameters) != 1 || req.ResponseParameters[0] != attrShare {
		t.Fatalf("responseParameters = %v", req.ResponseParameters)
	}
}

func TestResponseIndicatesAttributeNotDisplayable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		resp *mqscResponse
		want bool
	}{
		{
			name: "command message",
			resp: &mqscResponse{
				CommandResponse: []commandResponseItem{{
					CompletionCode: 2,
					Message:        []string{"MQWB0120E: Attribute SHARE is not valid."},
				}},
			},
			want: true,
		},
		{
			name: "rest error explanation",
			resp: &mqscResponse{
				Error: []restErrorItem{{
					Message:     "Bad request",
					Explanation: "MQWB0120E: response parameter not supported",
				}},
			},
			want: true,
		},
		{
			name: "unrelated failure",
			resp: &mqscResponse{
				CommandResponse: []commandResponseItem{{
					CompletionCode: 2,
					Message:        []string{"AMQ8147E: IBM MQ object APP.MISSING not found."},
				}},
			},
			want: false,
		},
		{
			name: "success",
			resp: &mqscResponse{
				CommandResponse: []commandResponseItem{{
					CompletionCode: 0,
					Parameters:     map[string]any{"maxdepth": "5000"},
				}},
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := responseIndicatesAttributeNotDisplayable(tc.resp); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestErrIndicatesAttributeNotDisplayable(t *testing.T) {
	t.Parallel()
	if !errIndicatesAttributeNotDisplayable(&mqadmin.TerminalError{
		Message: `mqweb returned HTTP 400: {"error":[{"msgId":"MQWB0120E"}]}`,
	}) {
		t.Fatal("expected HTTP 400 MQWB0120E terminal error to indicate not displayable")
	}
	if errIndicatesAttributeNotDisplayable(&mqadmin.TerminalError{
		Message: "mqweb returned HTTP 400: invalid queue name",
	}) {
		t.Fatal("unexpected match for unrelated bad request")
	}
}

func TestQueueLocalDefineOnlyCandidates(t *testing.T) {
	t.Parallel()
	want := []string{attrShare, attrDefopts, attrBothresh, attrBoqname, attrUsage, attrMaxMsgLen}
	if len(QueueLocalDefineOnlyCandidates) != len(want) {
		t.Fatalf("candidates = %v, want %v", QueueLocalDefineOnlyCandidates, want)
	}
	for i, attr := range want {
		if QueueLocalDefineOnlyCandidates[i] != attr {
			t.Fatalf("candidates[%d] = %q, want %q", i, QueueLocalDefineOnlyCandidates[i], attr)
		}
	}
	if len(queueLocalProbedDisplayCandidates) != len(QueueLocalDefineOnlyCandidates) {
		t.Fatalf("probed candidates = %v, want all define-only candidates", queueLocalProbedDisplayCandidates)
	}
}
