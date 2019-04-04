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
	"gonder/cmd"
	"gonder/models"
	"io"
	"log"
	"os"
	"runtime"
)

func main() {
	l, err := os.OpenFile(models.WorkDir("log/main.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening log file: %v", err)
	}
	defer l.Close()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(io.MultiWriter(l, os.Stdout))

	runtime.GOMAXPROCS(runtime.NumCPU())

	cmd.Run()
}
