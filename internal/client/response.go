package client

import (
	"encoding/json"
	"io"
	"net/http"
)

// Response http.Response wrapper that can interpret the body more.
type Response struct {
	*http.Response
}

// BodyBytes return http.Response body as a []byte.
func (r *Response) BodyBytes() ([]byte, error) {
	return io.ReadAll(r.Body)
}

// JSONMarshallBody unmarshall http.Response body as json to passed target.
func (r *Response) JSONMarshallBody(target interface{}) error {
	bodyBytes, err := r.BodyBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyBytes, target)
}
