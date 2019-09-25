package bindata

import (
	"io/ioutil"
	"os"
)

func ReadFileOrAsset(filename string) ([]byte, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return ioutil.ReadFile(filename)
	}
	return Asset(filename)
}
