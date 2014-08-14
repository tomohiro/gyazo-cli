package main

import (
	"bytes"
	"encoding/json"
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

	// Default endpoint of the Gyazo upload API.
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

// Set the endpoint of Gyazo upload API.
func init() {
	if os.Getenv("GYAZO_SERVER_URL") != "" {
		endpoint = os.Getenv("GYAZO_SERVER_URL")
	} else if os.Getenv("GYAZO_ACCESS_TOKEN") != "" {
		endpoint = "https://upload.gyazo.com/api/upload"
	}
}

// Image is response object of Gyazo upload API.
type Image struct {
	ID           string `json:"image_id"`
	PermalinkURL string `json:"permalink_url"`
	ThumbURL     string `json:"thumb_url"`
	URL          string `json:"url"`
	Type         string `json:"type"`
}

// Upload image to Gyazo server.
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
	defer content.Close()

	// Create multipart/form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("imagedata", filename)
	if err != nil {
		exitCode = 1
		return
	}

	if _, err = io.Copy(part, content); err != nil {
		exitCode = 1
		return
	}

	if os.Getenv("GYAZO_ACCESS_TOKEN") != "" {
		writer.WriteField("access_token", os.Getenv("GYAZO_ACCESS_TOKEN"))
	}

	err = writer.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create multipart/data: %s\n", err)
		exitCode = 1
		return
	}

	res, err := http.Post(endpoint, writer.FormDataContentType(), body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to upload: %s\n", err)
		exitCode = 1
		return
	}
	defer res.Body.Close()

	if os.Getenv("GYAZO_ACCESS_TOKEN") != "" {
		image := Image{}
		if err = json.NewDecoder(res.Body).Decode(&image); err != nil {
			fmt.Fprintf(os.Stderr, "Response error: %s\n", err)
			exitCode = 1
			return
		}

		fmt.Println(image.PermalinkURL)
	} else {
		url, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Response error: %s\n", err)
			exitCode = 1
			return
		}

		fmt.Println(string(url))
	}
}
