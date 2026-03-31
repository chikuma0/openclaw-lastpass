package protocol

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecodeRequest(t *testing.T) {
	t.Parallel()

	req, err := DecodeRequest(strings.NewReader(`{"protocolVersion":1,"provider":"lastpass","ids":["providers/openai/apiKey"]}`))
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}

	if req.ProtocolVersion != 1 {
		t.Fatalf("ProtocolVersion = %d, want 1", req.ProtocolVersion)
	}
	if req.Provider != "lastpass" {
		t.Fatalf("Provider = %q, want lastpass", req.Provider)
	}
	if len(req.IDs) != 1 || req.IDs[0] != "providers/openai/apiKey" {
		t.Fatalf("IDs = %#v, want one ID", req.IDs)
	}
}

func TestDecodeRequestRejectsTrailingContent(t *testing.T) {
	t.Parallel()

	if _, err := DecodeRequest(strings.NewReader(`{"protocolVersion":1,"provider":"lastpass","ids":[]} {}`)); err == nil {
		t.Fatal("DecodeRequest() error = nil, want trailing content error")
	}
}

func TestRequestValidateRejectsBadProvider(t *testing.T) {
	t.Parallel()

	req := &Request{
		ProtocolVersion: 1,
		Provider:        "something-else",
		IDs:             []string{"providers/openai/apiKey"},
	}

	if err := req.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want provider error")
	}
}

func TestWriteResponse(t *testing.T) {
	t.Parallel()

	response := NewResponse()
	response.Values["providers/openai/apiKey"] = "secret"
	response.Errors["providers/missing/apiKey"] = ErrorDetail{Message: "mapping not found"}

	var buf bytes.Buffer
	if err := WriteResponse(&buf, response); err != nil {
		t.Fatalf("WriteResponse() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"protocolVersion":1`) {
		t.Fatalf("output = %q, want protocolVersion", output)
	}
	if !strings.Contains(output, `"providers/openai/apiKey":"secret"`) {
		t.Fatalf("output = %q, want values content", output)
	}
	if !strings.Contains(output, `"message":"mapping not found"`) {
		t.Fatalf("output = %q, want errors content", output)
	}
}
