package main

import (
	"bytes"
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

	"github.com/mitchellh/go-homedir"
	"github.com/skratchdot/open-golang/open"
	"github.com/tomohiro/go-gyazo/gyazo"
	"github.com/urfave/cli/v2"
)

const (
	// ExitCodeOK represents succeed exit status
	ExitCodeOK int = 0

	// ExitCodeError represents failed exit status
	ExitCodeError int = 1
)

var (
	// exitCode to terminate.
	exitCode = ExitCodeOK

	// Default endpoint.
	endpoint = "http://upload.gyazo.com/upload.cgi"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	app := &cli.App{
		Name:    "gyazo-cli",
		Usage:   "Gyazo command-line uploader",
		Action:  upload,
		Version: Version,
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Tomohiro Taira",
				Email: "tomohiro.t@gmail.com",
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		exitCode = ExitCodeError
	}

	return exitCode
}

// Set the endpoint of self hosting Gyazo server URL.
func init() {
	if os.Getenv("GYAZO_SERVER_URL") != "" {
		endpoint = os.Getenv("GYAZO_SERVER_URL")
	}
}

// Upload a new image to a Gyazo from the specified local image file.
func upload(c *cli.Context) error {
	var filename string
	var url string
	var err error

	filename = c.Args().First()

	if filename == "" {
		filename, err = takeScreenshot()
		if err != nil {
			fmt.Fprint(os.Stderr, "Failed to take a screenshot\n")
			return err
		}
	}

	if !supportedMimetype(filename) {
		fmt.Fprint(os.Stderr, "Failed to upload: unsupported file type\n")
		return err
	}

	// Open and load the content from an image.
	content, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer content.Close()

	if os.Getenv("GYAZO_ACCESS_TOKEN") != "" {
		gyazo, _ := gyazo.NewClient(os.Getenv("GYAZO_ACCESS_TOKEN"))
		image, err := gyazo.Upload(content)
		if err != nil {
			fmt.Fprint(os.Stderr, "Failed to upload via API request\n")
			return err
		}
		url = image.PermalinkURL
	} else {
		// Create multipart/form-data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("imagedata", filename)
		if err != nil {
			return err
		}

		if _, err = io.Copy(part, content); err != nil {
			return err
		}

		writer.WriteField("id", gyazoID())
		err = writer.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create multipart/data: %s\n", err)
			return err
		}

		res, err := http.Post(endpoint, writer.FormDataContentType(), body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to upload: %s\n", err)
			return err
		}
		defer res.Body.Close()

		url, err = imageURL(res)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid response: %s\n", err)
			return err
		}
	}

	fmt.Println(url)
	err = open.Run(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open by default browser: %s\n", err)
	}
	return err
}

// takeScreenshot takes a screenshot and then returns it file path.
func takeScreenshot() (string, error) {
	var err error
	path := fmt.Sprintf("/tmp/image_upload%d.png", os.Getpid())

	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("screencapture", "-i", path).Run()
	case "linux":
		err = exec.Command("import", path).Run()
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

	id := r.Header.Get("X-Gyazo-Id")
	if err := storeGyazoID(id); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to store Gyazo ID: %s\n", err)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return url, err
	}

	url = string(body)

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
