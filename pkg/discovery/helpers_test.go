package discovery

import (
	"testing"
)

func TestSimpleJobCreator_EmptyIDErrors(t *testing.T) {
	create := SimpleJobCreator("svc", func(interface{}) string { return "" }, func(interface{}) string { return "proj" })
	_, err := create(struct{}{}, "res")
	if err == nil {
		t.Fatalf("expected error for empty resource ID")
	}
}


