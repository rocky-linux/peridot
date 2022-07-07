package xmlrpc

import "encoding/xml"

// Response is the basic parsed object of the XML-RPC response body.
// While it's not convenient to use this object directly - it contains all the information needed to unmarshal into other data-types.
type Response struct {
	Params []ResponseParam `xml:"params>param"`
	Fault  *ResponseFault  `xml:"fault,omitempty"`
}

// NewResponse creates a Response object from XML body.
// It relies on XML Unmarshaler and if it fails - error is returned.
func NewResponse(body []byte) (*Response, error) {

	response := &Response{}
	if err := xml.Unmarshal(body, response); err != nil {
		return nil, err
	}

	return response, nil
}

// ResponseParam encapsulates a nested parameter value
type ResponseParam struct {
	Value ResponseValue `xml:"value"`
}

// ResponseValue encapsulates one of the data types for each parameter.
// Only one field should be set.
type ResponseValue struct {
	Array    []*ResponseValue        `xml:"array>data>value"`
	Struct   []*ResponseStructMember `xml:"struct>member"`
	String   string                  `xml:"string"`
	Int      string                  `xml:"int"`
	Int4     string                  `xml:"i4"`
	Double   string                  `xml:"double"`
	Boolean  string                  `xml:"boolean"`
	DateTime string                  `xml:"dateTime.iso8601"`
	Base64   string                  `xml:"base64"`
}

// ResponseStructMember contains name-value pair of the struct
type ResponseStructMember struct {
	Name  string        `xml:"name"`
	Value ResponseValue `xml:"value"`
}

// ResponseFault wraps around failure
type ResponseFault struct {
	Value ResponseValue `xml:"value"`
}
