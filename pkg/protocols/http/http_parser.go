package http

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

	pb "github.com/dangai-8s/kook/api/v1"
	types "github.com/dangai-8s/kook/pkg/ebpf"
	"github.com/dangai-8s/kook/pkg/protocols"
)

type HTTPRequest struct {
	Record *http.Request
}

var _ protocols.ProtoRequest = &HTTPRequest{}

// ProtoType implements ProtoRequest
func (*HTTPRequest) ProtoType() types.ProtocolType { return types.HTTP }

// Protobuf implements ProtoRequest
func (r *HTTPRequest) Protobuf() *pb.Request {
	return &pb.Request{
		Record: &pb.Request_Http{
			Http: &pb.HTTPRequest{
				Protocol: r.Record.Proto,
				Method:   r.Record.Method,
				Url:      r.Record.RequestURI,
				Headers:  toProtobufHeader(r.Record.Header),
			},
		},
	}
}

func toProtobufHeader(header http.Header) []*pb.HTTPHeader {
	if header == nil {
		return nil
	}
	res := make([]*pb.HTTPHeader, 0, len(header))
	for k, vs := range header {
		for _, v := range vs {
			res = append(res, &pb.HTTPHeader{Key: k, Value: v})
		}
	}
	return res
}

type HTTPResponse struct {
	Record *http.Response
}

var _ protocols.ProtoResponse = &HTTPResponse{}

// ProtoType implements ProtoResponse
func (*HTTPResponse) ProtoType() types.ProtocolType { return types.HTTP }

// Protobuf implements ProtoResponse
func (r *HTTPResponse) Protobuf() *pb.Response {
	return &pb.Response{
		Record: &pb.Response_Http{
			Http: &pb.HTTPResponse{
				Protocol: r.Record.Proto,
				Code:     uint32(r.Record.StatusCode),
				Headers:  toProtobufHeader(r.Record.Header),
			},
		},
	}
}

const h1ReaderBufSize = 4096

type HTTPParser struct {
	reqReader  *bufio.Reader
	respReader *bufio.Reader
}

func NewHTTPParser() *HTTPParser {
	return &HTTPParser{
		reqReader:  bufio.NewReaderSize(nil, h1ReaderBufSize),
		respReader: bufio.NewReaderSize(nil, h1ReaderBufSize),
	}
}

var _ protocols.ProtoParser = &HTTPParser{}

// GetProtoType implements ProtoParser
func (p *HTTPParser) ProtoType() types.ProtocolType {
	return types.HTTP
}

// ParseRequest implements ProtoParser
func (p *HTTPParser) ParseRequest(_ *types.SockKey, msg []byte) ([]protocols.ProtoRequest, error) {
	r := p.reqReader
	br := bytes.NewReader(msg)
	r.Reset(br)

	req, err := http.ReadRequest(r)
	if err != nil {
		return nil, err
	}
	req.Body.Close()

	if !validMethod(req) {
		return nil, fmt.Errorf("invalid http method. got: %s", req.Method)
	}

	return []protocols.ProtoRequest{&HTTPRequest{Record: req}}, nil
}

func validMethod(req *http.Request) bool {
	switch req.Method {
	case http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
		http.MethodConnect,
		http.MethodTrace:
		return true
	default:
		return false
	}
}

func (p *HTTPParser) EnableInferRequest() bool {
	return true
}

// ParseResponse implements ProtoParser
func (p *HTTPParser) ParseResponse(_ *types.SockKey, msg []byte) ([]protocols.ProtoResponse, error) {
	r := p.respReader
	br := bytes.NewReader(msg)
	r.Reset(br)

	resp, err := http.ReadResponse(r, nil)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return []protocols.ProtoResponse{&HTTPResponse{Record: resp}}, nil
}

func (p *HTTPParser) EnableInferResponse() bool {
	return true
}
