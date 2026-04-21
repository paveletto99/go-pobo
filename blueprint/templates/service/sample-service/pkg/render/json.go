package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-multierror"
)

func (r *Renderer) RenderJSON(w http.ResponseWriter, code int, data interface{}) {
	if data == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if code >= 200 && code < 300 {
			fmt.Fprint(w, `{"ok":true}`)
			return
		}
		fmt.Fprintf(w, `{"error":"%s"}`, http.StatusText(code))
		return
	}

	if typ, ok := data.(*multierror.Error); ok {
		errs := typ.WrappedErrors()
		msgs := make([]string, 0, len(errs))
		for _, err := range errs {
			msgs = append(msgs, err.Error())
		}
		data = struct {
			Errors []string `json:"errors,omitempty"`
		}{Errors: msgs}
	}

	if typ, ok := data.(error); ok {
		data = struct {
			Error string `json:"error,omitempty"`
		}{Error: typ.Error()}
	}

	buf := r.pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer r.pool.Put(buf)

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"%s"}`, http.StatusText(http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = buf.WriteTo(w)
}
