package mqrest

import (
	"errors"
	"testing"

	"github.com/konradheimel/kurator/internal/mqadmin"
)

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
