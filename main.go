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
	// exitCode to terminate
	exitCode = 0

	// endpoint of the Gyazo upload API endpoint.
	endpoint = "http://upload.gyazo.com/upload.cgi"
)

const usageText = `Gyazo command-line uploader

EXAMPLE:

  $ gyazo foo.png
  $ gyazo ~/Downloads/bar.jpg`

func main() {
	app := cli.NewApp()
	app.Name = "gyazo"
	app.Version = Version
	app.Usage = usageText
	app.Author = "Tomohiro TAIRA"
	app.Email = "tomohiro.t@gmail.com"
	app.Action = upload
	app.Run(os.Args)

	os.Exit(exitCode)
}

func upload(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Try `gyazo --help` for more information")
		exitCode = 1
		return
	}

	filename := c.Args().First()
	content, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open and read %s\n", filename)
		return
	}

	// Create multipart/form-data
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

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

	res, err := http.Post(endpoint, w.FormDataContentType(), &buf)
	defer res.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to upload: %s", err)
		exitCode = 1
		return
	}

	url, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Response error: %s", err)
		exitCode = 1
		return
	}

	fmt.Println(url)
}
