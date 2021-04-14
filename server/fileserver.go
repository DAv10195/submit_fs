package fileserver

import (
	"github.com/spf13/viper"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)


func uploadFile(res http.ResponseWriter, req *http.Request) {
	// error and status to be returned by default after every execution of the function.
	var (
		status int
		err  error
	)
	defer func() {
		if err != nil {
			logger.WithError(err)
			http.Error(res, err.Error(), status)
		}
	}()
	// max memory: 20^32 mb
	if err = req.ParseMultipartForm(32 << 20); nil != err {
		status = http.StatusInternalServerError
		logger.Error("Cannot get file from request - Multi part form issue")
		logger.WithError(err)
		return
	}
	logger.Info("Starting the uploading")
	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			var infile multipart.File
			if infile, err = hdr.Open(); err != nil {
				logger.Error("Cannot open file from request body")
				logger.WithError(err)
				status = http.StatusInternalServerError
				return
			}
			// get the path in which we want to store the file from the request URL.
			path := filepath.Dir(req.URL.Path)
			// Create the path in the file server if not exist.
			err := os.MkdirAll(filepath.Join(viper.GetString("file-sever-path"), path), 0755)
			if err != nil {
				logger.Error("Error creating user directories")
				logger.WithError(err)
				status = http.StatusInternalServerError
				return
			}
			fullFilePath := filepath.Join(viper.GetString("file-sever-path"), filepath.Join(path, hdr.Filename))
			logger.Info("file path is:" + fullFilePath)
			var outfile *os.File
			if outfile, err = os.Create(fullFilePath); nil != err {
				logger.Error("Error creating user file in file server")
				logger.WithError(err)
				status = http.StatusInternalServerError
				return
			}
			// Copy to destination folder
			var written int64
			if written, err = io.Copy(outfile, infile); nil != err {
				logger.Error("Error copying user file in file server")
				logger.WithError(err)
				status = http.StatusInternalServerError
				return
			}
			//report the client about the uploaded file name and size.
			res.Write([]byte("uploaded file:" + hdr.Filename + ";length:" + strconv.Itoa(int(written))))
		}
	}
	logger.Info("Finished uploading")
}

