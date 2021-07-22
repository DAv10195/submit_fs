package server

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"net/http"
	"os"
)

func NewFileServer(router *mux.Router, tlsConf *tls.Config) *http.Server {
	return &http.Server{
		Addr: fmt.Sprintf(":%s", viper.GetString("file-server-port")),
		Handler:      router,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
		TLSConfig:    tlsConf,
	}
}

func InitRouters(r *mux.Router, path string) *mux.Router {
	r.PathPrefix("/").HandlerFunc(getUploadHandler(path).ServeHTTP).Methods(http.MethodPost)
	r.PathPrefix("/").HandlerFunc(getDownloadHandler(path).ServeHTTP).Methods(http.MethodGet)
	r.PathPrefix("/").HandlerFunc(getDeleteHandler(path).ServeHTTP).Methods(http.MethodDelete)

	return r
}

func InitFolders(){
	err := os.MkdirAll(viper.GetString("file-sever-path"), 0755)
	if err != nil {
		logger.WithError(err)
		return
	}
}

func GetTlsConfig(certFilePath, keyFilePath string) (*tls.Config, error) {
	if certFilePath != "" && keyFilePath != "" {
		cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
		if err != nil {
			return nil, err
		}
		return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
	}
	return nil, nil
}
