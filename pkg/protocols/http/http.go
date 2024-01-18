package http

import (
	"container/list"

	pb "github.com/dangai-8s/kook/api/v1/kook"
	"github.com/dangai-8s/kook/pkg/protocols"
)

type HTTPMatcher struct {
	reqQueue  *list.List
	respQueue *list.List
}

func NewHTTPMatcher() *HTTPMatcher {
	return &HTTPMatcher{
		reqQueue:  list.New(),
		respQueue: list.New(),
	}
}

var _ protocols.ProtoMatcher = &HTTPMatcher{}

// MatchRequest implements ProtoMatcher
func (m *HTTPMatcher) MatchRequest(req *protocols.Request) *pb.ProtoMessage {
	resp := m.findResp(req)
	if resp == nil {
		m.reqQueue.PushBack(req)
		return nil
	}
	return protocols.ProtoMessage(req, resp)
}

func (m *HTTPMatcher) findResp(req *protocols.Request) *protocols.Response {
	for e := m.respQueue.Front(); e != nil; e = e.Next() {
		if e.Value.(*protocols.Response).SockKey == req.SockKey {
			return m.respQueue.Remove(e).(*protocols.Response)
		}
	}
	return nil
}

// MatchResponse implements ProtoMatcher
func (m *HTTPMatcher) MatchResponse(resp *protocols.Response) *pb.ProtoMessage {
	req := m.findReq(resp)
	if req == nil {
		m.respQueue.PushBack(resp)
		return nil
	}

	return protocols.ProtoMessage(req, resp)
}

func (m *HTTPMatcher) findReq(resp *protocols.Response) *protocols.Request {
	for e := m.reqQueue.Front(); e != nil; e = e.Next() {
		if e.Value.(*protocols.Request).SockKey == resp.SockKey {
			return m.reqQueue.Remove(e).(*protocols.Request)
		}
	}
	return nil
}
