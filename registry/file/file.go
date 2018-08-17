package file

import (
	"io/ioutil"
	"log"
	"os"
	"time"
)

//file data
type filedata struct {
	path    string
	content string
	mtime   time.Time
}

// Example
// file := &filedata{path:"/home/zjj/routes.txt"}
// err := readFile(file)
func readFile(file *filedata) error {
	finfo, err := os.Stat(file.path)
	if err != nil {
		log.Println("[ERROR] Cannot read file stats() from ", file.path)
		return err
	}

	lastmtime := finfo.ModTime()
	if file.mtime != lastmtime {
		data, err := ioutil.ReadFile(file.path)
		if err != nil {
			log.Println("[ERROR] Cannot read file data from ", file.path)
			return err
		}
		file.content = string(data)
		file.mtime = lastmtime
	}
	return nil
}
