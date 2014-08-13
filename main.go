package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
)

var (
	// os.Exit forcely kills process.
	// So let me share this global variable to terminate at the last.
	exitCode = 0

	// endpoint of the Gyazo upload API endpoint.
	endpoint = "http://upload.gyazo.com/upload.cgi"
)

func main() {
	app := cli.NewApp()
	app.Name = "gyazo"
	app.Version = Version
	app.Usage = `Gyazo command-line uploader

EXAMPLE:

  $ gyazo foo.png
  $ gyazo ~/Downloads/bar.jpg`
	app.Author = "Tomohiro TAIRA"
	app.Email = "tomohiro.t@gmail.com"
	app.Action = doMain
	app.Run(os.Args)
	os.Exit(exitCode)
}

func doMain(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Println("gyazo: try `gyazo --help` for more information")
		exitCode = 1
		return
	}

	filename := c.Args().First()
	content, err := os.Open(filename)
	if err != nil {
		os.Stderr.WriteString("failed to open and read " + filename)
		exitCode = 1
		return
	}

	// Create multipart/form-data
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	f, err := w.CreateFormFile("imagedata", filename)
	if err != nil {
		exitCode = 1
		return
	}

	if _, err = io.Copy(f, content); err != nil {
		exitCode = 1
		return
	}

	w.Close()

	res, err := http.Post(endpoint, w.FormDataContentType(), &b)
	defer res.Body.Close()
	if err != nil {
		os.Stderr.WriteString("failed to upload")
		exitCode = 1
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		exitCode = 1
		return
	}

	println(string(body))
}
