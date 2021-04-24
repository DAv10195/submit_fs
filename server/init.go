package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"net/http"
	"os"
)

func NewFileServer(router *mux.Router) *http.Server {
	return &http.Server{
		Addr: fmt.Sprintf(":%s", viper.GetString("file-server-port")),
		Handler:      router,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}
}

func InitRouters(r *mux.Router, path string) *mux.Router {
	handler := http.FileServer(http.Dir(path))
	r.PathPrefix("/").HandlerFunc(handler.ServeHTTP).Methods(http.MethodGet)
	r.PathPrefix("/").HandlerFunc(getUploadHandler(path).ServeHTTP).Methods(http.MethodPost)
	return r
}

func InitFolders(){
	err := os.MkdirAll(viper.GetString("file-sever-path"), 0755)
	if err != nil {
		logger.WithError(err)
		return
	}
}
