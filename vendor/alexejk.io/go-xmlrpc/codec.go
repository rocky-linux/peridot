package xmlrpc

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"net/url"
	"sync"
)

const defaultUserAgent = "alexejk.io/go-xmlrpc"

// Codec implements methods required by rpc.ClientCodec
// In this implementation Codec is the one performing actual RPC requests with http.Client.
type Codec struct {
	endpoint      *url.URL
	httpClient    *http.Client
	customHeaders map[string]string

	mutex sync.Mutex
	// contains completed but not processed responses by sequence ID
	pending map[uint64]*rpcCall

	// Current in-flight response
	response *Response
	encoder  Encoder
	decoder  Decoder

	// presents completed requests by sequence ID
	ready chan uint64

	userAgent string
}

type rpcCall struct {
	Seq           uint64
	ServiceMethod string
	httpResponse  *http.Response
}

// NewCodec creates a new Codec bound to provided endpoint.
// Provided client will be used to perform RPC requests.
func NewCodec(endpoint *url.URL, httpClient *http.Client) *Codec {
	return &Codec{
		endpoint:   endpoint,
		httpClient: httpClient,
		encoder:    &StdEncoder{},
		decoder:    &StdDecoder{},

		pending:  make(map[uint64]*rpcCall),
		response: nil,
		ready:    make(chan uint64),

		userAgent: defaultUserAgent,
	}
}

func (c *Codec) WriteRequest(req *rpc.Request, args interface{}) error {

	bodyBuffer := new(bytes.Buffer)
	err := c.encoder.Encode(bodyBuffer, req.ServiceMethod, args)
	if err != nil {
		return err
	}

	httpRequest, err := http.NewRequest("POST", c.endpoint.String(), bodyBuffer)
	if err != nil {
		return err
	}

	httpRequest.Header.Set("Content-Type", "text/xml")
	httpRequest.Header.Set("User-Agent", c.userAgent)

	// Apply customer headers if set, this allows overwriting static default headers
	for key, value := range c.customHeaders {
		httpRequest.Header.Set(key, value)
	}

	httpRequest.Header.Set("Content-Length", fmt.Sprintf("%d", bodyBuffer.Len()))

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	c.pending[req.Seq] = &rpcCall{
		Seq:           req.Seq,
		ServiceMethod: req.ServiceMethod,
		httpResponse:  httpResponse,
	}
	c.mutex.Unlock()

	c.ready <- req.Seq

	return nil
}

func (c *Codec) ReadResponseHeader(resp *rpc.Response) error {

	seq := <-c.ready

	c.mutex.Lock()
	call := c.pending[seq]
	delete(c.pending, seq)
	c.mutex.Unlock()

	resp.Seq = call.Seq
	resp.ServiceMethod = call.ServiceMethod

	r := call.httpResponse

	defer r.Body.Close()

	if r.StatusCode < 200 || r.StatusCode >= 300 {
		resp.Error = fmt.Sprintf("bad response code: %d", r.StatusCode)
		return nil
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		resp.Error = err.Error()
		return nil
	}

	decodableResponse, err := NewResponse(body)
	if err != nil {
		resp.Error = err.Error()
		return nil
	}

	// Return response Fault already a this stage
	if err := c.decoder.DecodeFault(decodableResponse); err != nil {
		resp.Error = err.Error()
		return nil
	}

	c.response = decodableResponse

	return nil
}
func (c *Codec) ReadResponseBody(v interface{}) error {

	if v == nil {
		return nil
	}

	if c.response == nil {
		return errors.New("no in-flight response found")
	}

	return c.decoder.Decode(c.response, v)
}

func (c *Codec) Close() error {

	c.httpClient.CloseIdleConnections()

	return nil
}
