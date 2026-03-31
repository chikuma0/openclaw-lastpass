package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const Version1 = 1

type Request struct {
	ProtocolVersion int      `json:"protocolVersion"`
	Provider        string   `json:"provider"`
	IDs             []string `json:"ids"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

type Response struct {
	ProtocolVersion int                    `json:"protocolVersion"`
	Values          map[string]string      `json:"values"`
	Errors          map[string]ErrorDetail `json:"errors,omitempty"`
}

func NewResponse() *Response {
	return &Response{
		ProtocolVersion: Version1,
		Values:          make(map[string]string),
		Errors:          make(map[string]ErrorDetail),
	}
}

func DecodeRequest(r io.Reader) (*Request, error) {
	var req Request
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&req); err != nil {
		return nil, err
	}

	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != nil && !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected trailing JSON content")
	}
	if len(trailing) > 0 {
		return nil, errors.New("unexpected trailing JSON content")
	}
	return &req, nil
}

func (r *Request) Validate() error {
	if r.ProtocolVersion != Version1 {
		return fmt.Errorf("unsupported protocolVersion %d", r.ProtocolVersion)
	}

	if strings.TrimSpace(r.Provider) != "lastpass" {
		return fmt.Errorf("unsupported provider %q", r.Provider)
	}

	for _, id := range r.IDs {
		if strings.TrimSpace(id) == "" {
			return errors.New("ids must not contain empty values")
		}
	}

	return nil
}

func WriteResponse(w io.Writer, response *Response) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(response)
}
