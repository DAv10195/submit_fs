package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const logHttpErrFormat = "error serving http request for %s"

type Response struct {
	Message	string	`json:"message"`
}

func (e *Response) String() string {
	errRespBytes, _ := json.Marshal(e)
	return string(errRespBytes)
}

func writeResponse(w http.ResponseWriter, r *http.Request, httpStatus int, response *Response) {
	w.WriteHeader(httpStatus)
	if _, err := w.Write([]byte(response.String())); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	}
}

func writeStrErrResp(w http.ResponseWriter, r *http.Request, httpStatus int, str string) {
	err := fmt.Errorf(str)
	logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	writeResponse(w, r, httpStatus, &Response{err.Error()})
}
