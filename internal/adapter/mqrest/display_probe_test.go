package mqrest

import (
	"testing"
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

func TestQueueLocalDefineOnlyCandidatesIncludesShare(t *testing.T) {
	t.Parallel()
	found := false
	for _, attr := range QueueLocalDefineOnlyCandidates {
		if attr == attrShare {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected share in candidates, got %v", QueueLocalDefineOnlyCandidates)
	}
}
