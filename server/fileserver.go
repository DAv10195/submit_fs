package server

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
		defer func() {
			if err != nil {
				writeResponse(res, req, status, &Response{err.Error()})
				return
			}
		}()
		path := filepath.Dir(req.URL.Path)
		filePath := filepath.Join(fsPath,req.URL.String())
		if req.Header.Get("Content-type") != multipartform {
			// copy the body of the request to the file.
			// in this case the url will be used for the file name.
			// example:  a/b/c.txt in the request will form a file named c.txt in a/b and its content
			// will be the body of the request.
			err := os.MkdirAll(filepath.Join(fsPath, path), 0755)
			if err != nil {
				logger.WithError(err).Error("Error creating user directories")
				status = http.StatusInternalServerError
				return
			}
			//check if the file exist. if yes delete it first.
			// TODO: check david if its safe for us.
			if _, err := os.Stat(filePath); err != nil {
				err := os.Remove(filePath)
				if err != nil {
					logger.WithError(err).Error("Error removing existing file and replacing it")
					status = http.StatusInternalServerError
					return
				}
			}
			out, err := os.Create(filePath)
			if err != nil {
				logger.WithError(err).Error("Error creating user file (raw data from body)")
				status = http.StatusInternalServerError
				return
			}
			defer func() {
				err = out.Close()
				if err != nil {
					writeResponse(res, req, status, &Response{err.Error()})
					return
				}
			}()
			_, err = io.Copy(out, req.Body)
			if err != nil {
				logger.WithError(err).Error("Error copying the request body to file (raw data from body)")
				status = http.StatusInternalServerError
				return
			}
			return
		}
		res.Header().Set("Content-Type", "application/json")
		// max memory: 20^32 mb
		if err = req.ParseMultipartForm(32 << 20); nil != err {
			status = http.StatusInternalServerError
			logger.WithError(err).Error("Cannot get file from request - Multi part form parsing issue")
			return
		}
		logger.Info("Starting the uploading")
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
				err := os.MkdirAll(filepath.Join(fsPath, path), 0755)
				if err != nil {
					logger.WithError(err).Error("Error creating user directories")
					status = http.StatusInternalServerError
					return
				}
				fullFilePath := filepath.Join(fsPath, path, hdr.Filename)
				logger.Info("file path is:" + fullFilePath)
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
		logger.Info("Finished uploading")
	})
}
