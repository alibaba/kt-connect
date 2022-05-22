package util

import (
	"github.com/rs/zerolog/log"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

const logFilePrefix = "kt-"

var BackgroundLogger = io.Discard

func PrepareLogger(enable bool) {
	if !enable {
		return
	}
	if tmpFile, err := ioutil.TempFile(os.TempDir(), logFilePrefix); err != nil {
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

// Write log text
func (f FileWriter) Write(p []byte) (n int, err error) {
	return f.file.Write(p)
}

// CleanBackgroundLogs delete expired log files
func CleanBackgroundLogs() {
	files, err := ioutil.ReadDir(os.TempDir())
	if err != nil {
		log.Warn().Msgf("Failed to list background logs")
		return
	}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), logFilePrefix) && isExpired(f) {
			err = os.Remove(path.Join(os.TempDir(), f.Name()))
			if err != nil {
				log.Debug().Err(err).Msgf("Failed to removed expired log %s", f.Name())
			} else {
				log.Debug().Msgf("Removed expired log %s", f.Name())
			}
		}
	}
}

// isExpired check whether file haven't been modified over 24 hours
func isExpired(info fs.FileInfo) bool {
	return info.ModTime().Unix() < time.Now().Unix() - (3600 * 24)
}
