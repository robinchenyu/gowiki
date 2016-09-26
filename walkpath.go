package main

import (
	"fmt"
	// "io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func checkerr(e error, msg string) {
	if e != nil {
		panic(e)
	}
}

var chin = make(chan string)

func scanfiles(path string, info os.FileInfo, err error) error {
	checkerr(err, "scanfiles failed ")

	// fmt.Println("path: ", path, "fileinfo ", info.Name(), info.IsDir(), info.Size())
	if strings.HasSuffix(info.Name(), ".md") {
		unix_path := strings.Replace(path, "\\", "/", -1)
		chin <- fmt.Sprintf("[%s](%s)\n\n", info.Name(), unix_path[:len(path)-3])
	}
	return nil
}

func main1() {
	go func(ch chan string) {
		var data string
		fd, err := os.Create("index.md")
		checkerr(err, "create index.md failed! ")

		defer fd.Close()

		for {
			select {
			case data = <-ch:
				if data == "EOF" {
					break
				}
				fmt.Fprintf(fd, data)
			}
		}
	}(chin)

	err := filepath.Walk(".", scanfiles)
	checkerr(err, "walk file failed! ")
	fmt.Println("finish!")
	chin <- "EOF"
}
