package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

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
	var filename string
	var err error

	filename = c.Args().First()

	if filename == "" {
		filename, err = takeScreenshot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to take a screenshot: %s\n", err)
			exitCode = 1
			return
		}
	}

	if !supportedMimetype(filename) {
		fmt.Fprint(os.Stderr, "Failed to upload: unsupported file type\n")
		exitCode = 1
		return
	}

	// Open and load the content from an image.
	content, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open and read: %s\n", filename)
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

// takeScreenshot takes a screenshot and then returns it file path.
func takeScreenshot() (string, error) {
	var err error
	path := fmt.Sprintf("/tmp/image_upload%d.png", os.Getpid())

	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("screencapture", "-i", path).Run()
	case "linux":
		if _, err := exec.LookPath("import"); err == nil {
			err = exec.Command("import", path).Run()
		} else {
			return "", fmt.Errorf("'import' command not found")
		}
	case "windows":
		err = errors.New("unsupported os")
	}

	return path, err
}

// supportedMimetype returns result of checked mimetype.
func supportedMimetype(f string) bool {
	t := mime.TypeByExtension(filepath.Ext(f))
	res, _ := regexp.MatchString("image", t)
	return res
}

// imageURL returns url of uploaded image.
func imageURL(r *http.Response) (string, error) {
	var url string
	var err error
	if os.Getenv("GYAZO_ACCESS_TOKEN") != "" {
		image := Image{}
		if err = json.NewDecoder(r.Body).Decode(&image); err != nil {
			return url, err
		}

		url = image.PermalinkURL
	} else {
		id := r.Header.Get("X-Gyazo-Id")
		if err = storeGyazoID(id); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to store Gyazo ID: %s\n", err)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return url, err
		}

		url = string(body)
	}

	return url, err
}

// gyazoID returns Gyazo ID from stored file.
func gyazoID() string {
	var id string
	body, err := ioutil.ReadFile(gyazoIDPath())
	if err != nil {
		return id
	}

	id = string(body)
	return id
}

// storeGyazoID stores Gyazo ID to file.
func storeGyazoID(id string) error {
	var err error
	if id == "" {
		return err
	}

	path := gyazoIDPath()

	dir := filepath.Dir(path)
	_, err = os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(path)
	if err == nil {
		newpath := fmt.Sprintf("%s_%s.bak", id, time.Now().Format("20060102150405"))
		err = os.Rename(path, newpath)
		if err != nil {
			return err
		}
	}

	buf := bytes.NewBufferString(id)
	err = ioutil.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return err
}

// gyazoIDPath returns path of Gyazo ID file on local filesystem.
func gyazoIDPath() string {
	homedir, _ := homedir.Dir()

	var path string
	switch runtime.GOOS {
	case "darwin":
		path = fmt.Sprintf("%s/Library/Gyazo/id", homedir)
	case "linux":
		path = fmt.Sprintf("%s/.gyazo.id", homedir)
	case "windows":
		path = fmt.Sprintf("%s\\Gyazo\\id.txt", os.Getenv("APPDATA"))
	}

	return path
}
