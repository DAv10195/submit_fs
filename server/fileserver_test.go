package server

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
	tempDir := filepath.Join(os.TempDir(),"files")
	defer os.Remove(tempDir)
	dummyTestPath := filepath.Join(tempDir, "david.txt")
	targzPath := filepath.Join(tempDir, "nikita", "nikita.tar.gz")
	// create the files for the test.
	if err := os.MkdirAll(filepath.Join(tempDir,"nikita"),0755); err != nil{
		t.Fatalf("Failed to create folders for test: %v", err)
	}
	pathForTar := filepath.Join(tempDir, "david")
	if err := os.MkdirAll(pathForTar,0755); err != nil {
		t.Fatalf("Failed to create tar gz folders for test: %v", err)
	}
	f,err := os.Create(filepath.Join(tempDir,"david","file.txt"))
	if err != nil {
		t.Fatalf("Failed to create file for tar gz: %v", err)
	}
	_, err = f.Write([]byte("nikita"))
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}

	dummyFile,err := os.Create(dummyTestPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer dummyFile.Close()
	err = ioutil.WriteFile(dummyTestPath, []byte("david"), 0755)
	if err != nil {
		t.Fatalf("Unable to write file: %v", err)
	}
	//create the targz file
	targz, err := os.Create(targzPath)
	if err != nil {
		t.Fatalf("Unable to create file: %v", err)
	}
	err = Compress(pathForTar,targz)
	if err != nil {
		t.Fatalf("Unable to compress file: %v", err)
	}

	defer os.RemoveAll(tempDir)
	testCases := []struct{
		name	string
		method	string
		path	string
		file    *os.File
		status	int
		isAdmin	bool
	}{
		{
			"test upload file to server as annonymous",
			http.MethodPost,
			dummyTestPath,
			dummyFile,
			http.StatusUnauthorized,
			false,
		},
		{
			"test get file from server as annonymous",
			http.MethodGet,
			dummyTestPath,
			dummyFile,
			http.StatusUnauthorized,
			false,
		},
		{
			"test upload file to server as admin",
			http.MethodPost,
			dummyTestPath,
			dummyFile,
			http.StatusAccepted,
			true,
		},
		{
			"test upload folder to server as admin",
			http.MethodPost,
			targzPath,
			targz,
			http.StatusAccepted,
			true,
		},
		{
			"test upload text from body as admin",
			http.MethodPost,
			dummyTestPath,
			dummyFile,
			http.StatusAccepted,
			true,
		},
		{
			"test download file from server as admin",
			http.MethodGet,
			dummyTestPath,
			dummyFile,
			http.StatusOK,
			true,
		},
		{
			"test download folder from server as admin",
			http.MethodGet,
			"/",
			nil,
			http.StatusOK,
			true,
		},
	}

	router := mux.NewRouter()
	tmpDir := filepath.Join(os.TempDir(), "test-fs")
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("enable to create test fs root: %v", err)
	}

	viper.Set("file-server-path", tmpDir)
	err = InitFsEncryption()
	if err != nil {
		t.Fatalf("error creating encryption key for : %v",  err)
	}
	defer os.Remove(filepath.Join(tmpDir, "submit_file_server.key"))
	password, err := fsEncryption.Encrypt(DefPass)
	if err != nil {
		t.Fatalf("error creating password encryption for test: %v", err)
	}
	viper.Set("password", password)
	viper.Set("user", DefUser)
	router = InitRouters(router, tmpDir)
	router.Use(AuthenticationMiddleware)
	body := &bytes.Buffer{}
	for _, testCase := range testCases {
		w := httptest.NewRecorder()
		var r *http.Request
		var testCaseErr error
		if !t.Run(testCase.name, func (t *testing.T) {
			if testCase.method == http.MethodPost {
				if testCase.name == "test upload text from body as admin"{
					body = bytes.NewBuffer([]byte("nikita"))
					// put string in the body and send the request. a/b/c.txt
					r, err = http.NewRequest(testCase.method,"/" + filepath.Base(testCase.path), body)
					r.SetBasicAuth(DefUser, DefPass)
					router.ServeHTTP(w, r)
					if w.Code != testCase.status {
						testCaseErr = fmt.Errorf("test case [ %s ] produced status code %d instead of the expected %d status code", testCase.name, w.Code, testCase.status)
						t.FailNow()
					}
					return
				}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", filepath.Base(testCase.file.Name()))
				var written int64 = 0
				var file *os.File
				file, err = os.Open(testCase.path)
				if err != nil {
					testCaseErr = fmt.Errorf("error Opening test case [ %s ]: %v", testCase.name, err)
					t.FailNow()
				}
				written, err = io.Copy(part, file)
				logger.Print(written)
				if err != nil {
					testCaseErr = fmt.Errorf("error getting files for test case [ %s ]: %v", testCase.name, err)
					t.FailNow()
				}
				err = writer.Close()
				if err != nil {
					testCaseErr = fmt.Errorf("error closing files for test case [ %s ]: %v", testCase.name, err)
					t.FailNow()
				}
				r, err = http.NewRequest(testCase.method, "/", body)
				if testCase.name == "test upload folder to server as admin"{
					q := r.URL.Query()
					q.Add("isFolder", "true")
					r.URL.RawQuery = q.Encode()
				}
				r.Header.Add("Content-Type", writer.FormDataContentType())
			} else {
				if testCase.name == "test download folder from server as admin"{
					r, err = http.NewRequest(testCase.method, "/" , nil)
				}else {
					r, err = http.NewRequest(testCase.method, "/" + filepath.Base(testCase.path), nil)
				}
			}
			if err != nil {
				testCaseErr = fmt.Errorf("error creating http request for test case [ %s ]: %v", testCase.name, err)
				t.FailNow()
			}
			if testCase.isAdmin {
				r.SetBasicAuth(DefUser, DefPass)
			}
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
