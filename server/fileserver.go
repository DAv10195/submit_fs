package server

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	commons "github.com/DAv10195/submit_commons"
)

func getUploadHandler(fsPath string) http.Handler {
	return http.HandlerFunc(func (res http.ResponseWriter, req *http.Request) {
		// error and status to be returned by default after every execution of the function.
		var (
			status int
			err  error
			totalBytesWritten int64
			uploadedFileNames []string
		)
		res.Header().Set("Content-Type", "application/json")
		defer func() {
			if err != nil {
				writeStrErrResp(res, req, status, err.Error())
			}
		}()
		path := req.URL.Path
		filePath := filepath.Join(fsPath,req.URL.String())
		reqType := req.Header.Get("Content-type")
		if !strings.Contains(reqType, multipartform){
			// copy the body of the request to the file.
			// in this case the url will be used for the file name.
			// example:  a/b/c.txt in the request will form a file named c.txt in a/b and its content
			// will be the body of the request.

			err = os.MkdirAll(filepath.Dir(filePath), 0755)
			if err != nil {
				logger.WithError(err).Error("Error creating user directories")
				status = http.StatusInternalServerError
				return
			}
			//check if the file exist. if yes delete it first.
			logger.Debug(fmt.Printf("Checking if file exist for request %s", req.URL.Path))
			if _, err = os.Stat(filePath); err == nil {
					err = os.Remove(filePath)
					if err != nil {
						logger.WithError(err).Error("Error removing existing file and replacing it")
						status = http.StatusInternalServerError
						return
					}
			} else if !os.IsNotExist(err){
				logger.WithError(err).Error("Error uploading the file - file already exist and cannot be deleten")
				status = http.StatusInternalServerError
				return
			}
			var out *os.File
			logger.Debug(fmt.Printf("Creating file for request %s", req.URL.Path))
			out, err = os.Create(filePath)
			if err != nil {
				logger.WithError(err).Error("Error creating user file (raw data from body)")
				status = http.StatusInternalServerError
				return
			}
			defer func() {
				err = out.Close()
				if err != nil {
					status = http.StatusInternalServerError
					return
				}
			}()
			_, err = io.Copy(out, req.Body)
			if err != nil {
				logger.WithError(err).Error("Error copying the request body to file (raw data from body)")
				status = http.StatusInternalServerError
				return
			}
			//write response
			logger.Debug(fmt.Printf("writing the body from request %s to file", req.URL.Path))
			writeResponse(res, req, http.StatusAccepted, &Response{fmt.Sprintf("Uploaded Files: %v. Total Bytes Written: %v", uploadedFileNames, totalBytesWritten)})
			return
		}
		// max memory: 20^32 mb
		logger.Debug(fmt.Printf("Starting to parse multi part form for request: %s", req.URL.Path))
		if err = req.ParseMultipartForm(32 << 20); nil != err {
			status = http.StatusInternalServerError
			logger.WithError(err).Error("Cannot get file from request - Multi part form parsing issue")
			return
		}
		logger.Debug(fmt.Printf("Getting file headers from multi part from for request: %s", req.URL.Path))
		for _, fheaders := range req.MultipartForm.File {
			for _, hdr := range fheaders {
				var infile multipart.File
				if infile, err = hdr.Open(); err != nil {
					logger.WithError(err).Error("Cannot open file from request body")
					status = http.StatusInternalServerError
					return
				}
				// get the path in which we want to store the file from the request URL.
				// Create the path in the file server if not exist.
				path = filepath.Dir(path)
				err = os.MkdirAll(filepath.Join(fsPath, path), 0755)
				if err != nil {
					logger.WithError(err).Error("Error creating user directories")
					status = http.StatusInternalServerError
					return
				}
				fullFilePath := filepath.Join(fsPath, path, hdr.Filename)
				logger.Debug(fmt.Printf("Handling the file %s for request %s", hdr.Filename, req.URL.Path))
				var outfile *os.File
				if outfile, err = os.Create(fullFilePath); nil != err {
					logger.WithError(err).Error("Error creating user file in file server")
					status = http.StatusInternalServerError
					return
				}
				// Copy to destination folder
				var written int64
				if written, err = io.Copy(outfile, infile); nil != err {
					logger.WithError(err).Error("Error copying user file in file server")
					status = http.StatusInternalServerError
					return
				}
				uploadedFileNames = append(uploadedFileNames, hdr.Filename)
				totalBytesWritten += written
			}
		}
		writeResponse(res, req, http.StatusAccepted, &Response{fmt.Sprintf("Uploaded Files: %v. Total Bytes Written: %v", uploadedFileNames, totalBytesWritten)})
	})
}

func getDownloadHandler(fsPath string) http.Handler {
	return http.HandlerFunc(func (res http.ResponseWriter, req *http.Request) {
		var (
			status int
			err  error
		)
		defer func() {
			if err != nil {
				writeStrErrResp(res, req, status, err.Error())
				return
			}
		}()
		path := filepath.Join(fsPath,req.URL.String())
		//first check if its folder or file.
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				logger.WithError(err).Error("file/folder does not exist")
				status = http.StatusNotFound
				return
			}
			logger.WithError(err).Error("error getting the file/folder")
			status = http.StatusInternalServerError
			return
		}
		if info.IsDir() {
			path = req.URL.String()
			f := filepath.Join(os.TempDir(), commons.GenerateUniqueId())
			fullPathToTar := fmt.Sprintf("%s.tar.gz", f)
			var tarFile *os.File
			tarFile, err = os.Create(fullPathToTar)
			if err != nil {
				logger.WithError(err).Error("Failed to create the tar gz file")
				status = http.StatusInternalServerError
				return
			}
			err = Compress(filepath.Join(fsPath,path), tarFile)
			if err != nil {
				logger.WithError(err).Error("Failed to compress the folder")
				status = http.StatusInternalServerError
				return
			}
			// put the compressed file into the response.
			logger.Debug(fmt.Printf("Downloading the file %s,",fullPathToTar))
			http.ServeFile(res,req,fullPathToTar)
		} else {
			logger.Debug(fmt.Printf("Downloading the file %s,",path))
			http.ServeFile(res,req,path)
		}
	})
}