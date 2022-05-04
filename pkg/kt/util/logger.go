package util

import (
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"os"
)

var BackgroundLogger = io.Discard

func PrepareLogger(enable bool) {
	if !enable {
		return
	}
	if tmpFile, err := ioutil.TempFile(os.TempDir(), "kt-"); err != nil {
		log.Warn().Err(err).Msgf("Cannot create background log file")
	} else {
		log.Debug().Msgf("Background task log to %s", tmpFile.Name())
		FixFileOwner(tmpFile.Name())
		BackgroundLogger = FileWriter{tmpFile}
	}
}

type FileWriter struct {
	file *os.File
}

func (f FileWriter) Write(p []byte) (n int, err error) {
	return f.file.Write(p)
}
