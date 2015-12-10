// Copyright (C) 2015 Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// NewRespBodyFromString creates an io.ReadCloser from a string that is suitable for use as an
// http response body.
func NewRespBodyFromString(body string) io.ReadCloser {
	return &dummyReadCloser{strings.NewReader(body)}
}

type dummyReadCloser struct {
	body io.ReadSeeker
}

func (d *dummyReadCloser) Read(p []byte) (n int, err error) {
	n, err = d.body.Read(p)
	if err == io.EOF {
		d.body.Seek(0, 0)
	}
	return n, err
}

func (d *dummyReadCloser) Close() error {
	return nil
}

func NewResponse(body string, status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       NewRespBodyFromString(body),
		Header:     http.Header{},
	}
}

type Foo struct {
	Message string `json:"message,omitempty"`
}

func TestDecodeResponse(t *testing.T) {
	resp := NewResponse(`{"message":"Project not found"}`, http.StatusNotFound)
	var f *Foo
	DecodeResponse(resp, &f)
	if f.Message != "Project not found" {
		t.Fatalf("Invalid decode JSON")
	}
}
