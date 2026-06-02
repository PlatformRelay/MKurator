package mqrest

import (
	"errors"
	"testing"

	"github.com/konradheimel/kurator/internal/mqadmin"
)

func TestMqscResponseOverallFailed_RESTErrorOnly(t *testing.T) {
	t.Parallel()
	resp := &mqscResponse{Error: []restErrorItem{{Message: "mqweb failure"}}}
	if !resp.overallFailed() {
		t.Fatal("expected failure when REST error present")
	}
}

func TestMqscResponseOverallFailed(t *testing.T) {
	t.Parallel()
	ok := &mqscResponse{OverallCompletionCode: 0, CommandResponse: []commandResponseItem{{CompletionCode: 0}}}
	if ok.overallFailed() {
		t.Fatal("expected success")
	}
	fail := &mqscResponse{
		OverallCompletionCode: 2,
		CommandResponse: []commandResponseItem{{
			CompletionCode: 2,
			Message:        []string{"AMQ8147E: IBM MQ object X not found."},
		}},
	}
	if !fail.overallFailed() {
		t.Fatal("expected failure")
	}
	if !fail.isObjectMissing() {
		t.Fatal("expected object missing")
	}
	err := fail.terminalError("display")
	var term *mqadmin.TerminalError
	if !errors.As(err, &term) {
		t.Fatalf("expected TerminalError, got %T", err)
	}
}

func TestMqscResponseFirstMessageVariants(t *testing.T) {
	t.Parallel()
	text := &mqscResponse{
		CommandResponse: []commandResponseItem{{
			CompletionCode: 2,
			Text:           []string{"text error"},
		}},
	}
	if text.firstMessage() != "text error" {
		t.Fatalf("firstMessage = %q", text.firstMessage())
	}
	rest := &mqscResponse{Error: []restErrorItem{{Message: "rest error"}}}
	if rest.firstMessage() != "rest error" {
		t.Fatalf("firstMessage = %q", rest.firstMessage())
	}
	empty := &mqscResponse{}
	if empty.firstMessage() != "unknown mqsc error" {
		t.Fatalf("firstMessage = %q", empty.firstMessage())
	}
}

func TestMqscResponseFirstObjectAttributes(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		resp := &mqscResponse{
			CommandResponse: []commandResponseItem{{
				CompletionCode: 0,
				Parameters:     map[string]any{"MaxDepth": 5000, "descr": "orders"},
			}},
		}
		attrs, err := resp.firstObjectAttributes()
		if err != nil || attrs["maxdepth"] != "5000" || attrs["descr"] != "orders" {
			t.Fatalf("attrs=%v err=%v", attrs, err)
		}
	})
	t.Run("empty response is not found", func(t *testing.T) {
		t.Parallel()
		_, err := (&mqscResponse{}).firstObjectAttributes()
		if !errors.Is(err, mqadmin.ErrNotFound) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("missing object", func(t *testing.T) {
		t.Parallel()
		resp := &mqscResponse{
			OverallCompletionCode: 2,
			CommandResponse: []commandResponseItem{{
				CompletionCode: 2,
				Message:        []string{"AMQ8147E: not found"},
			}},
		}
		_, err := resp.firstObjectAttributes()
		if !errors.Is(err, mqadmin.ErrNotFound) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("terminal on failed display", func(t *testing.T) {
		t.Parallel()
		resp := &mqscResponse{
			OverallCompletionCode: 2,
			CommandResponse: []commandResponseItem{{
				CompletionCode: 2,
				Message:        []string{"AMQ9999E: serious failure"},
			}},
		}
		_, err := resp.firstObjectAttributes()
		var term *mqadmin.TerminalError
		if !errors.As(err, &term) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestMqscResponseIsObjectMissing_AMQ8101(t *testing.T) {
	t.Parallel()
	resp := &mqscResponse{
		CommandResponse: []commandResponseItem{{
			Message: []string{"AMQ8101: object not found"},
		}},
	}
	if !resp.isObjectMissing() {
		t.Fatal("expected object missing")
	}
}
