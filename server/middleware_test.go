package fileserver

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)


func TestFileMiddleware(t *testing.T) {
	upFilePath := "/nikita/kogan/up.txt"
	downFilePath := "/david/abramov/down.txt"
	//fileToUpload := "up.txt"
	testCases := []struct{
		name	string
		method	string
		path	string
		status	int
		isAdmin	bool
	}{
		{
			"test upload file to server as annonymous",
			http.MethodPost,
			upFilePath,
			http.StatusUnauthorized,
			false,
		},
		{
			"test get file from server as annonymous",
			http.MethodGet,
			downFilePath,
			http.StatusUnauthorized,
			false,
		},
		{
			"test upload file to server as admin",
			http.MethodPost,
			upFilePath,
			http.StatusOK,
			true,
		},
		{
			"test get  file from server as admin",
			http.MethodGet,
			downFilePath,
			http.StatusOK,
			true,
		},
	}
	router := getRouterForTest(t)
	router.Use(AuthenticationMiddleware)
	for _, testCase := range testCases {
		var body []byte
		var testCaseErr error
		if !t.Run(testCase.name, func (t *testing.T) {
			r, err := http.NewRequest(testCase.method, testCase.path, bytes.NewBuffer(body))
			if err != nil {
				testCaseErr = fmt.Errorf("error creating http request for test case [ %s ]: %v", testCase.name, err)
				t.FailNow()
			}
			password := "admin"
			username := "admin"
			if testCase.isAdmin {
				r.SetBasicAuth(username, password)
			}
			router.Use(AuthenticationMiddleware)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			if w.Code != testCase.status {
				testCaseErr = fmt.Errorf("test case [ %s ] produced status code %d instead of the expected %d status code", testCase.name, w.Code, testCase.status)
				t.FailNow()
			}
		}) {
			t.Logf("error in test case [ %s ]: %v", testCase.name, testCaseErr)
		}
	}
}

func getRouterForTest(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/nikita/kogan/up.txt", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("\"message\": \"hello from download endpoint\"")); err != nil {
			t.FailNow()
		}
	})
	router.HandleFunc("/david/abramov/down.txt", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("\"message\": \"hello from upload endpoint\"")); err != nil {
			t.FailNow()
		}
	})
	return router
}
