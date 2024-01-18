package http

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	types "github.com/dangai-8s/kook/pkg/ebpf"
	"github.com/dangai-8s/kook/pkg/protocols"
	"golang.org/x/net/http2"
)

func TestHTTPRequestRecord_ProtoType(t *testing.T) {
	tests := []struct {
		name string
		want types.ProtocolType
	}{
		{
			name: "HTTP Request Record Protocol Type",
			want: types.HTTP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HTTPRequest{}
			if got := h.ProtoType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPRequestRecord.ProtoType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPResponseRecord_ProtoType(t *testing.T) {
	tests := []struct {
		name string
		want types.ProtocolType
	}{
		{
			name: "HTTP Response Record Protocol Type",
			want: types.HTTP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HTTPResponse{}
			if got := h.ProtoType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPResponseRecord.ProtoType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPParser_ProtoType(t *testing.T) {
	tests := []struct {
		name string
		want types.ProtocolType
	}{
		{
			name: "HTTP Parser Protocol Type",
			want: types.HTTP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewHTTPParser()
			if got := p.ProtoType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPParser.GetProtoType() = %v, want %v", got, tt.want)
			}
		})
	}
}

type HTTPParser_ParseRequest_Test struct {
	name    string
	req     *http.Request
	wantErr bool
}

func (t *HTTPParser_ParseRequest_Test) args() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	t.req.Write(buf)

	len := buf.Len()
	if len > 4096 {
		len = 4096
	}
	return buf.Bytes()[:len]
}

func (t *HTTPParser_ParseRequest_Test) want() protocols.ProtoRequest {
	return &HTTPRequest{Record: t.req}
}

func (t *HTTPParser_ParseRequest_Test) equal(got protocols.ProtoRequest) bool {
	h1rr, ok := got.(*HTTPRequest)
	if !ok {
		return false
	}
	wantBytes := make([]byte, 0, 4096)
	t.req.Write(bytes.NewBuffer(wantBytes))
	gotBytes := make([]byte, 0, 4096)
	h1rr.Record.Write(bytes.NewBuffer(gotBytes))

	return bytes.Equal(wantBytes, gotBytes)
}

func TestHTTPParser_ParseRequest(t *testing.T) {
	tests := []HTTPParser_ParseRequest_Test{
		{
			name: "Short Header Get Request",
			req: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://keploy.io/test/url", nil)
				req.Header.Add("User-Agent", "test-clinet/1.1")
				return req
			}(),
			wantErr: false,
		},
		{
			name: "Long Header Get Request",
			req: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://keploy.io/test/url", nil)
				req.Header.Add("User-Agent", "test-clinet/1.1")
				req.Header.Add("Long-Cookie", strings.Repeat("1234567890", 500))
				return req
			}(),
			wantErr: true,
		},
		{
			name: "Short Header Short Body Post Request",
			req: func() *http.Request {
				body := bytes.NewReader([]byte(strings.Repeat("1234567890", 10)))
				req, _ := http.NewRequest(http.MethodPost, "http://keploy.io/test/url", body)
				req.Header.Add("User-Agent", "test-clinet/1.1")
				return req
			}(),
			wantErr: false,
		},
		{
			name: "Short Header Long Body Post Request",
			req: func() *http.Request {
				body := bytes.NewReader([]byte(strings.Repeat("1234567890", 500)))
				req, _ := http.NewRequest(http.MethodPost, "http://keploy.io/test/url", body)
				req.Header.Add("User-Agent", "test-clinet/1.1")
				return req
			}(),
			wantErr: false,
		},
		{
			name: "Long Header Long Body Post Request",
			req: func() *http.Request {
				body := bytes.NewReader([]byte(strings.Repeat("1234567890", 500)))
				req, _ := http.NewRequest(http.MethodPost, "http://keploy.io/test/url", body)
				req.Header.Add("User-Agent", "test-clinet/1.1")
				req.Header.Add("Long-Cookie", strings.Repeat("1234567890", 500))
				return req
			}(),
			wantErr: true,
		},
		{
			name: "HTTP/2 Preface",
			req: func() *http.Request {
				req, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(http2.ClientPreface)))
				return req
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewHTTPParser()

			got, err := p.ParseRequest(nil, tt.args())
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPParser.ParseRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !tt.equal(got[0]) {
				t.Errorf("HTTPParser.ParseRequest() = %v, want %v", got, tt.want())
			}
		})
	}
}

func Test_validMethod(t *testing.T) {
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// valid
		{
			name: "Method GET",
			args: args{req: &http.Request{Method: http.MethodGet}},
			want: true,
		},
		{
			name: "Method Post",
			args: args{req: &http.Request{Method: http.MethodPost}},
			want: true,
		},
		{
			name: "Method Put",
			args: args{req: &http.Request{Method: http.MethodPut}},
			want: true,
		},
		{
			name: "Method Delete",
			args: args{req: &http.Request{Method: http.MethodDelete}},
			want: true,
		},
		{
			name: "Method Patch",
			args: args{req: &http.Request{Method: http.MethodPatch}},
			want: true,
		},
		{
			name: "Method Head",
			args: args{req: &http.Request{Method: http.MethodHead}},
			want: true,
		},
		{
			name: "Method Options",
			args: args{req: &http.Request{Method: http.MethodOptions}},
			want: true,
		},
		{
			name: "Method Connect",
			args: args{req: &http.Request{Method: http.MethodConnect}},
			want: true,
		},
		{
			name: "Method Trace",
			args: args{req: &http.Request{Method: http.MethodTrace}},
			want: true,
		},

		// invalid
		{
			name: "Method PRI",
			args: args{req: &http.Request{Method: "PRI"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validMethod(tt.args.req); got != tt.want {
				t.Errorf("isValidMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

type HTTPParser_ParseResponse_Test struct {
	name    string
	resp    *http.Response
	wantErr bool
}

func (t *HTTPParser_ParseResponse_Test) args() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	t.resp.Write(buf)

	len := buf.Len()
	if len > 4096 {
		len = 4096
	}
	return buf.Bytes()[:len]
}

func (t *HTTPParser_ParseResponse_Test) want() protocols.ProtoResponse {
	return &HTTPResponse{Record: t.resp}
}

func (t *HTTPParser_ParseResponse_Test) equal(got protocols.ProtoResponse) bool {
	h1rr, ok := got.(*HTTPResponse)
	if !ok {
		return false
	}
	wantBytes := make([]byte, 0, 4096)
	t.resp.Write(bytes.NewBuffer(wantBytes))
	gotBytes := make([]byte, 0, 4096)
	h1rr.Record.Write(bytes.NewBuffer(gotBytes))

	return bytes.Equal(wantBytes, gotBytes)
}

func TestHTTPParser_ParseResponse(t *testing.T) {
	tests := []HTTPParser_ParseResponse_Test{
		{
			name: "Short Header Short Body",
			resp: func() *http.Response {
				resp := &http.Response{}
				resp.Proto = "HTTP/1.1"
				resp.StatusCode = http.StatusOK
				resp.Header = make(http.Header)
				resp.Header.Add("User-Agent", "test-clinet/1.1")
				resp.Body = io.NopCloser(bytes.NewReader([]byte(strings.Repeat("1234567890", 10))))

				return resp
			}(),
			wantErr: false,
		},
		{
			name: "Short Header Long Body",
			resp: func() *http.Response {
				resp := &http.Response{}
				resp.Proto = "HTTP/1.1"
				resp.StatusCode = http.StatusOK
				resp.Header = make(http.Header)
				resp.Header.Add("User-Agent", "test-clinet/1.1")
				resp.Body = io.NopCloser(bytes.NewReader([]byte(strings.Repeat("1234567890", 500))))
				return resp
			}(),
			wantErr: false,
		},
		{
			name: "Long Header Short Body",
			resp: func() *http.Response {
				resp := &http.Response{}
				resp.Proto = "HTTP/1.1"
				resp.StatusCode = http.StatusOK
				resp.Header = make(http.Header)
				resp.Header.Add("User-Agent", "test-clinet/1.1")
				resp.Header.Add("Long-Cookie", strings.Repeat("1234567890", 500))
				resp.Body = io.NopCloser(bytes.NewReader([]byte(strings.Repeat("1234567890", 10))))
				return resp
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewHTTPParser()

			got, err := p.ParseResponse(nil, tt.args())
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPParser.ParseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !tt.equal(got[0]) {
				t.Errorf("HTTPParser.ParseResponse() = %v, want %v", got, tt.want())
			}
		})
	}
}
