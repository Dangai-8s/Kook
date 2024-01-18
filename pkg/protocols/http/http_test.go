package http

import (
	"container/list"
	"net/http"
	"reflect"
	"testing"

	types "github.com/dangai-8s/kook/pkg/ebpf/types"
	"github.com/dangai-8s/kook/pkg/protocols"
)

func TestHTTPMatcher_MatchRequest(t *testing.T) {
	type fields struct {
		reqQueue  *list.List
		respQueue *list.List
	}
	type args struct {
		req *protocols.Request
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   bool
	}{
		{
			name: "Success to Find Response",
			args: args{
				req: &protocols.Request{
					SockKey: types.SockKey{Pid: 1},
					Record:  &HTTPRequest{&http.Request{}},
				},
			},
			fields: fields{
				reqQueue: list.New(),
				respQueue: func() *list.List {
					l := list.New()
					l.PushFront(&protocols.Response{
						SockKey: types.SockKey{Pid: 1},
						Record:  &HTTP1Response{&http.Response{}},
					})
					return l
				}(),
			},
			want: true,
		},
		{
			name: "Empty Response Queue",
			args: args{
				req: &protocols.Request{
					SockKey: types.SockKey{Pid: 1},
					Record:  &HTTPRequest{&http.Request{}},
				},
			},
			fields: fields{
				reqQueue:  list.New(),
				respQueue: list.New(),
			},
			want: false,
		},
		{
			name: "Fail to Find Response",
			args: args{
				req: &protocols.Request{
					SockKey: types.SockKey{Pid: 1},
					Record:  &HTTPRequest{&http.Request{}},
				},
			},
			fields: fields{
				reqQueue: list.New(),
				respQueue: func() *list.List {
					l := list.New()
					l.PushFront(&protocols.Response{
						SockKey: types.SockKey{Pid: 2},
						Record:  &HTTP1Response{&http.Response{}},
					})
					return l
				}(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HTTPMatcher{
				reqQueue:  tt.fields.reqQueue,
				respQueue: tt.fields.respQueue,
			}
			got := m.MatchRequest(tt.args.req)
			if (got != nil) != tt.want {
				t.Errorf("HTTPMatcher.MatchRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPMatcher_findResp(t *testing.T) {
	type fields struct {
		reqQueue  *list.List
		respQueue *list.List
	}
	type args struct {
		req *protocols.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *protocols.Response
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HTTPMatcher{
				reqQueue:  tt.fields.reqQueue,
				respQueue: tt.fields.respQueue,
			}
			if got := m.findResp(tt.args.req); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPMatcher.findResp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPMatcher_MatchResponse(t *testing.T) {
	type fields struct {
		reqQueue  *list.List
		respQueue *list.List
	}
	type args struct {
		resp *protocols.Response
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Success to Find Request",
			args: args{
				resp: &protocols.Response{
					SockKey: types.SockKey{Pid: 1},
					Record:  &HTTP1Response{&http.Response{}},
				},
			},
			fields: fields{
				reqQueue: func() *list.List {
					l := list.New()
					l.PushFront(&protocols.Request{
						SockKey: types.SockKey{Pid: 1},
						Record:  &HTTPRequest{&http.Request{}},
					})
					return l
				}(),
				respQueue: list.New(),
			},
			want: true,
		},
		{
			name: "Empty Request Queue",
			args: args{
				resp: &protocols.Response{
					SockKey: types.SockKey{Pid: 1},
					Record:  &HTTP1Response{&http.Response{}},
				},
			},
			fields: fields{
				reqQueue:  list.New(),
				respQueue: list.New(),
			},
			want: false,
		},
		{
			name: "Fail to Find Request",
			args: args{
				resp: &protocols.Response{
					SockKey: types.SockKey{Pid: 1},
					Record:  &HTTP1Response{&http.Response{}},
				},
			},
			fields: fields{
				reqQueue: func() *list.List {
					l := list.New()
					l.PushFront(&protocols.Request{
						SockKey: types.SockKey{Pid: 2},
						Record:  &HTTPRequest{&http.Request{}},
					})
					return l
				}(),
				respQueue: list.New(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HTTPMatcher{
				reqQueue:  tt.fields.reqQueue,
				respQueue: tt.fields.respQueue,
			}
			got := m.MatchResponse(tt.args.resp)
			if (got != nil) != tt.want {
				t.Errorf("HTTPMatcher.MatchResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPMatcher_findReq(t *testing.T) {
	type fields struct {
		reqQueue  *list.List
		respQueue *list.List
	}
	type args struct {
		resp *protocols.Response
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *protocols.Request
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HTTPMatcher{
				reqQueue:  tt.fields.reqQueue,
				respQueue: tt.fields.respQueue,
			}
			if got := m.findReq(tt.args.resp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPMatcher.findReq() = %v, want %v", got, tt.want)
			}
		})
	}
}
