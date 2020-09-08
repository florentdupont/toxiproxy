package toxics

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Shopify/toxiproxy/stream"
)

type HttpToxic struct{
	// Times in milliseconds
	Status int64 `json:"status"`
}


func (t *HttpToxic) ModifyResponse(resp *http.Response) {
	resp.Header.Set("Content-Type", "text/plain")

	resp.StatusCode = 500
	bodyContent := "Internal Server Error"

	switch t.Status {
	case 401:
		resp.StatusCode = 401
		bodyContent = "Unauthorized"
	case 403:
		resp.StatusCode = 403
		bodyContent = "Forbidden"
	case 404:
		resp.StatusCode = 404
		bodyContent = "Not Found"
	default:

	}
	resp.Status = bodyContent
	resp.Body = ioutil.NopCloser(strings.NewReader(bodyContent))
	resp.ContentLength = int64(len(bodyContent))
}

func (t *HttpToxic) Pipe(stub *ToxicStub) {
	buffer := bytes.NewBuffer(make([]byte, 0, 32*1024))
	writer := stream.NewChanWriter(stub.Output)
	reader := stream.NewChanReader(stub.Input)
	reader.SetInterrupt(stub.Interrupt)
	for {
		tee := io.TeeReader(reader, buffer)
		resp, err := http.ReadResponse(bufio.NewReader(tee), nil)
		if err == stream.ErrInterrupted {
			buffer.WriteTo(writer)
			return
		} else if err == io.EOF {
			stub.Close()
			return
		}
		if err != nil {
			buffer.WriteTo(writer)
		} else {
			t.ModifyResponse(resp)
			resp.Write(writer)
		}
		buffer.Reset()
	}
}

func init() {
	Register("http", new(HttpToxic))
}