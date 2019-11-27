Description
================================================================================

Gyazo command-line uploader


Usage
--------------------------------------------------------------------------------

```
$ gyazo-cli [PATH]
```

### Example

#### Take a screenshot and then upload

NOTE: this feature is not available on Windows. ([ImageMagick](http://www.imagemagick.org/script/index.php) is required for Linux users)

```
$ gyazo-cli
```

#### Uploading a specific image

```
$ gyazo-cli ~/Desktop/image.png
http://gyazo.com/f1380d79593d2aaa0fcd412511f3d3e5
```


Configuration
--------------------------------------------------------------------------------

### Use Gyazo API client token

Set the access token to environment variable like this:

```
export GYAZO_ACCESS_TOKEN="YOUR GYAZO API ACCESS TOKEN"
```


### Use self-hosted Gyazo server

Set the server URL to environment variable like this:

```
export GYAZO_SERVER_URL="http://my-gyazo.example.com"
```


Installation
--------------------------------------------------------------------------------

### Get the stable binary

Go to the [release page](https://github.com/tomohiro/gyazo-cli/releases) and download a zip file.


### go get

Install to `$GOPATH/bin`:

```
$ GO111MODULE=off go get -u github.com/tomohiro/gyazo-cli
```


Contributing
--------------------------------------------------------------------------------

See [CONTRIBUTING](CONTRIBUTING.md) guideline.


LICENSE
--------------------------------------------------------------------------------

&copy; 2014 - 2019 Tomohiro Taira.

This project is licensed under the MIT license. See [LICENSE](LICENSE) for details.
