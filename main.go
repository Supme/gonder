// Project Gonder.
// Author Supme
// Copyright Supme 2016
// License http://opensource.org/licenses/MIT MIT License
//
//  THE SOFTWARE AND DOCUMENTATION ARE PROVIDED "AS IS" WITHOUT WARRANTY OF
//  ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
//  IMPLIED WARRANTIES OF MERCHANTABILITY AND/OR FITNESS FOR A PARTICULAR
//  PURPOSE.
//
// Please see the License.txt file for more information.
//
package main

import (
	"github.com/supme/gonder/cmd"
	"github.com/supme/gonder/models"
	"io"
	"log"
	"os"
	"runtime"
)

func main() {
	l, err := os.OpenFile(models.FromRootDir("log/main.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening log file: %v", err)
	}
	defer l.Close()

	ml := io.MultiWriter(l, os.Stdout)

	log.SetFlags(3)
	log.SetOutput(ml)

	runtime.GOMAXPROCS(runtime.NumCPU())

	//	models.Config.Get()
	defer models.Config.Close()

	cmd.Run()
}
