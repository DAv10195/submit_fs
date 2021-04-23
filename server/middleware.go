package server

import (
	"github.com/spf13/viper"
	"net/http"
)

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok {
			logger.Error("failed fetching user and password from request")
			writeStrErrResp(w, req, http.StatusUnauthorized, unauthorized)
			return
		}
		decryptedPass, err := fsEncryption.Decrypt(viper.GetString("password"))
		if err != nil {
			logger.WithError(err).Error("decryption password error")
			writeStrErrResp(w, req, http.StatusUnauthorized, unauthorized)
			return
		}
		if viper.GetString("user") != username || password != decryptedPass {
			logger.Error("Auth error - please authenticate with valid creds")
			writeStrErrResp(w, req, http.StatusUnauthorized, unauthorized)
			return
		}
		next.ServeHTTP(w, req)
	})

}
