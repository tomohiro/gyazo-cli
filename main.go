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
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
	"github.com/skratchdot/open-golang/open"
)

var (
	// exitCode to terminate.
	exitCode = 0

	// Default endpoint.
	endpoint = "http://upload.gyazo.com/upload.cgi"
)

func main() {
	defer os.Exit(exitCode)

	app := cli.NewApp()
	app.Name = "gyazo"
	app.Version = Version
	app.Usage = "Gyazo command-line uploader"
	app.Author = "Tomohiro TAIRA"
	app.Email = "tomohiro.t@gmail.com"
	app.Action = upload
	app.Run(os.Args)
}

// Set the endpoint of Gyazo API.
func init() {
	if os.Getenv("GYAZO_SERVER_URL") != "" {
		endpoint = os.Getenv("GYAZO_SERVER_URL")
	} else if os.Getenv("GYAZO_ACCESS_TOKEN") != "" {
		endpoint = "https://upload.gyazo.com/api/upload"
	}
}

// Image represents a uploaded image on the Gyazo server.
//
// Gyazo API docs: https://gyazo.com/api/docs/image
type Image struct {
	ID           string `json:"image_id"`
	PermalinkURL string `json:"permalink_url"`
	ThumbURL     string `json:"thumb_url"`
	URL          string `json:"url"`
	Type         string `json:"type"`
}

// Upload a new image to a Gyazo server from the specified local image file.
//
// Gyazo API docs: https://gyazo.com/api/docs/image
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
	} else {
		writer.WriteField("id", gyazoID())
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

	url, err := imageURL(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Response error: %s\n", err)
		exitCode = 1
		return
	}

	fmt.Println(url)
	err = open.Run(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open by default browser: %s\n", err)
	}
}

// imageURL returns url of uploaded image.
func imageURL(r *http.Response) (string, error) {
	var url = ""
	if os.Getenv("GYAZO_ACCESS_TOKEN") != "" {
		image := Image{}
		if err := json.NewDecoder(r.Body).Decode(&image); err != nil {
			return url, err
		}

		url = image.PermalinkURL
	} else {
		id := r.Header.Get("X-Gyazo-Id")
		if err := saveGyazoID(id); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save Gyazo ID: %s\n", err)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return url, err
		}

		url = string(body)
	}

	return url, nil
}

// getID returns ID
func gyazoID() string {
	var id = ""

	fp, err := os.Open(gyazoIDPath())
	if err != nil {
		return id
	}
	defer fp.Close()

	body, err := ioutil.ReadAll(fp)
	if err != nil {
		return id
	}

	id = string(body)
	return id
}

// saveGyazoID stores Gyazo ID to file.
func saveGyazoID(id string) error {
	return nil
}

// gyazoIDPath returns path of Gyazo ID file on local filesystem.
func gyazoIDPath() string {
	homedir, _ := homedir.Dir()

	var path = ""
	switch runtime.GOOS {
	case "darwin":
		path = fmt.Sprintf("%s/Library/Gyazo/id", homedir)
	case "linux":
		path = fmt.Sprintf("%s/.gyazo.id", homedir)
	}

	return path
}
