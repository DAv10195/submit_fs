package fileserver

import (
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
		if username != adminUser || password != adminPass {
			logger.Error("Auth error - please authenticate with admin user")
			writeStrErrResp(w, req, http.StatusUnauthorized, wrongCreds)
			return
		}
		next.ServeHTTP(w, req)
	})

}

