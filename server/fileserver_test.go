package fileserver

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFileServerHandlers(t *testing.T) {
	testPath := "/david.txt"
	//upFilePath := filepath.Join(path.GetDefaultWorkDirPath(), testPath)
	//downFilePath := filepath.Join(path.GetDefaultWorkDirPath(),testPath)
	//InitFolders()
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
			testPath,
			http.StatusUnauthorized,
			false,
		},
		{
			"test get file from server as annonymous",
			http.MethodGet,
			testPath,
			http.StatusUnauthorized,
			false,
		},
		{
			"test upload file to server as admin",
			http.MethodPost,
			testPath,
			http.StatusOK,
			true,
		},
		{
			"test get  file from server as admin",
			http.MethodGet,
			testPath,
			http.StatusOK,
			true,
		},
	}
	body := &bytes.Buffer{}
	err := ioutil.WriteFile("david.txt", []byte("david"), 0755)
	if err != nil {
		fmt.Printf("Unable to write file: %v", err)
	}
	file, err := os.Open("david.txt")
	if err != nil {
		fmt.Printf("error creating file to upload : %v",  err)
		t.FailNow()
	}
	defer file.Close()
	router := mux.NewRouter()
	err = InitFsEncryption()
	if err != nil {
		fmt.Printf("error creating encryption key for : %v",  err)
	}
	router = InitRouters(router)
	router.Use(AuthenticationMiddleware)
	for _, testCase := range testCases {
		var r *http.Request
		var testCaseErr error
		if !t.Run(testCase.name, func (t *testing.T) {
			if testCase.method == http.MethodPost {
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
				io.Copy(part, file)
				writer.Close()
				r, err = http.NewRequest(testCase.method, testCase.path, body)
				r.Header.Add("Content-Type", writer.FormDataContentType())
			} else {
				r, err = http.NewRequest(testCase.method, testCase.path, nil)
			}
			if err != nil {
				testCaseErr = fmt.Errorf("error creating http request for test case [ %s ]: %v", testCase.name, err)
				t.FailNow()
			}

			password, err := Encrypt(DefAdminPass)
			if err != nil {
				testCaseErr = fmt.Errorf("error creating password encryption for test case [ %s ]: %v", testCase.name, err)
				t.FailNow()
			}
			username , err := Encrypt(DefAdminUser)
			if err != nil {
				testCaseErr = fmt.Errorf("error creating username encryption for test case [ %s ]: %v", testCase.name, err)
				t.FailNow()
			}
			viper.Set("admin-user", username)
			viper.Set("admin-password", password)
			if testCase.isAdmin {
				r.SetBasicAuth(DefAdminUser, DefAdminPass)
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
