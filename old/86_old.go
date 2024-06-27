// Copyright (c) 2016 Uber Technologies, Inc.
	// All behaviors will eventually contribute to this behavior's output.
	bt := BehaviorTester{Params: httpParams{Request: r}}
	v := bt.Param("behavior")
	switch v {
	case "raw", "json", "thrift":
		runEchoBehavior(&bt, v)
	default:
		bt.NewBehavior(BasicEntryBuilder).Skipf("unknown behavior %v", v)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package client

import (
	"net/http"

	"github.com/yarpc/yarpc-go/crossdock/client/behavior"
	"github.com/yarpc/yarpc-go/crossdock/client/echo"
)

// Start begins a blocking Crossdock client
func Start() {
	http.HandleFunc("/", behaviorRequestHandler)
	http.ListenAndServe(":8080", nil)
}

func behaviorRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "HEAD" {
		return
	}

	var s behavior.EntrySink
	behavior.Run(func() { dispatch(&s, httpParams{r}) })
	if err := s.WriteJSON(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func dispatch(s behavior.Sink, ps behavior.Params) {
	v := ps.Param("behavior")
	switch v {
	case "raw":
		echo.Raw(s, ps)
	case "json":
		echo.JSON(s, ps)
	case "thrift":
		echo.Thrift(s, ps)
	default:
		behavior.Skipf(s, "unknown behavior %q", v)
	}
}

// httpParams provides access to behavior parameters that are stored inside an
// HTTP request.
type httpParams struct {
	Request *http.Request
}

func (h httpParams) Param(name string) string {
	return h.Request.FormValue(name)
}
