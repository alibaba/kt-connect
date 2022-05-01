package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	tmpFile, _ := ioutil.TempFile(os.TempDir(), "kt-")
	fmt.Println(tmpFile.Name())
	tmpFile.Write([]byte("bac"))
	tmpFile.Write([]byte("ert"))
	tmpFile.Write([]byte("iou"))
	tmpFile.Write([]byte("nbx"))
}
