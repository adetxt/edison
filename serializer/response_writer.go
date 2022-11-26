package serializer

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type EdisonResponseWriter struct {
	io.Writer
	http.ResponseWriter
	Response *echo.Response
}

func (w *EdisonResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *EdisonResponseWriter) Write(b []byte) (int, error) {
	var i interface{}
	if err := json.Unmarshal(b, &i); err != nil {
		return 0, err
	}

	code := w.Response.Status
	isOK := code < 400

	res := map[string]interface{}{}

	if !isOK {
		res["status"] = "error"
		res["code"] = code
		res["error"] = strings.ToUpper(http.StatusText(code))
		res["message"] = i.(map[string]interface{})["message"]
	} else {
		res["status"] = "success"
		res["message"] = strings.ToUpper(http.StatusText(code))
		res["data"] = i
	}

	b, err := json.Marshal(res)
	if err != nil {
		return 0, err
	}

	return w.ResponseWriter.Write(b)
}

func (w *EdisonResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *EdisonResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
