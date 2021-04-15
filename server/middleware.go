package fileserver

import (
	"github.com/spf13/viper"
	"net/http"
)

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok {
			logger.Error("failed performing basic auth")
			writeStrErrResp(w, req, http.StatusUnauthorized, unauthorized)
			return
		}
		decryptedUser, err := Decrypt(viper.GetString("admin-user"))
		if err != nil {
			logger.Error("decryption user problem")
			writeStrErrResp(w, req, http.StatusUnauthorized, unauthorized)
			return
		}
		decryptedPass, err := Decrypt(viper.GetString("admin-password"))
		if err != nil {
			logger.Error("decryption password error ")
			writeStrErrResp(w, req, http.StatusUnauthorized, unauthorized)
			return
		}
		if username != decryptedUser || password != decryptedPass {
			logger.Error("Auth error - please authenticate with admin user / invalid creds")
			writeStrErrResp(w, req, http.StatusUnauthorized, wrongCreds)
			return
		}
		next.ServeHTTP(w, req)
	})

}

