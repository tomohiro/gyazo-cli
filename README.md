Gyazo CLI
================================================================================

Gyazo command-line uploader


Usage
--------------------------------------------------------------------------------

### Take a screenshot and then upload

NOTE: this feature is available on OS X only.

```sh
$ gyazo
```


### Upload specify image

```sh
$ gyazo ~/Desktop/image.png
http://gyazo.com/f1380d79593d2aaa0fcd412511f3d3e5
```


Configuration
--------------------------------------------------------------------------------

### Use Gyazo API client token

Set the access token to environment variable like this:

```sh
export GYAZO_ACCESS_TOKEN="YOUR GYAZO API ACCESS TOKEN"
```


### Use self-hosted Gyazo server

Set the access token to environment variable like this:

```sh
export GYAZO_SERVER_URL="http://my-gyazo.example.com"
```


Installation
--------------------------------------------------------------------------------

### Get the stable binary

Go to the [release page](https://github.com/Tomohiro/gyazo-cli/releases) and download a zip file.


### go get

Install to `$GOPATH/bin`:

```sh
$ go get -d github.com/Tomohiro/gyazo-cli
$ cd $GOPATH/src/github.com/Tomohiro/gyazo-cli
$ make install
```


LICENSE
--------------------------------------------------------------------------------

&copy; 2014 Tomohiro TAIRA.
This project is licensed under the MIT license.
See LICENSE for details.
