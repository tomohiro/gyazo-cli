// +build !linux

package main

import (
	"fmt"
)

func selectArea() (string, error) {
	return "", fmt.Errorf(`selectArea is not supported other than Linux`)
}
