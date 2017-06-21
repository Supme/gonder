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
package api

import (
	"encoding/json"
	"github.com/nfnt/resize"
	"github.com/supme/gonder/models"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var filemanagerRootPath string

type info struct {
	Path       string     `json:"Path"`
	Name       string     `json:"Name"`
	Filename   string     `json:"Filename"`
	FileType   string     `json:"File Type"`
	Preview    string     `json:"Preview"`
	Protected  int        `json:"Protected"`
	Properties properties `json:"Properties"`
	Error      string     `json:"Error"`
	Code       int        `json:"Code"`
	OldPath    string     `json:"Old Path"`
	OldName    string     `json:"Old Name"`
	NewPath    string     `json:"New Path"`
	NewName    string     `json:"New Name"`
}

type properties struct {
	DateCreated  string `json:"Date Created"`
	DateModified string `json:"Date Modified"`
	Filemtime    string `json:"filemtype"`
	Height       string `json:"Height"`
	Width        string `json:"Width"`
	Size         string `json:"Size"`
}

func filemanager(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte

	if auth.Right("filemanager") {
		var mode, path, name, oldName, newName, height, width string
		if r.Method == "GET" {
			if err = r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if r.Form["mode"] != nil {
				mode = r.Form["mode"][0]
			}
			if r.Form["path"] != nil {
				path = r.Form["path"][0]
			}
			if r.Form["name"] != nil {
				name = r.Form["name"][0]
			}
			if r.Form["old"] != nil {
				oldName = r.Form["old"][0]
			}
			if r.Form["new"] != nil {
				newName = r.Form["new"][0]
			}
			if r.Form["height"] != nil {
				height = r.Form["height"][0]
			}
			if r.Form["width"] != nil {
				width = r.Form["width"][0]
			}

			if mode == "download" {
				var n string
				var d []byte

				if width != "" && height != "" {
					// Download resized
					d = filemanagerResize(path, width, height)
				} else {
					// Download file
					n, d = filemanagerDownload(path)
					w.Header().Set("Content-Disposition", "attachment; filename='"+n+"'")
				}
				w.Write(d)
			} else {
				js, err = json.Marshal(filemanagerAction(mode, path, name, oldName, newName))
				if err != nil {
					apilog.Println(err)
				}
			}
		} else {
			if err = r.ParseMultipartForm(32 << 20); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if r.Form["mode"] != nil {
				mode = r.Form["mode"][0]
			}
			if mode == "add" {
				file, head, err := r.FormFile("newfile")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				js, err = json.Marshal(filemanagerAdd(r.Form["currentpath"][0], head.Filename, file))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	} else {
		js, err = json.Marshal(info{Error: "Filemanager forbidden", Code: 1})
		if err != nil {
			apilog.Println(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func filemanagerAction(mode string, path string, name string, oldName string, newName string) interface{} {
	filemanagerRootPath = "."
	switch mode {
	case "getinfo":
		return filemanagerGetInfo(path)
	case "getfolder":
		return filemanagerGetFolder(path)
	case "delete":
		return filemanagerDelete(path)
	case "addfolder":
		return filemanagerMkDir(path, name)
	case "rename":
		return filemanagerRename(oldName, newName)
	}

	return info{Error: "Mode not defined", Code: 1}
}

func filemanagerAdd(path string, name string, file multipart.File) info {
	out, err := os.Create(models.FromRootDir(filemanagerRootPath + path + name))
	if err != nil {
		return info{Error: "Can not create file", Code: 1}
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return info{Error: "Can not write file", Code: 1}
	}

	return info{Path: path, Name: name, Error: "No error", Code: 0}
}

func filemanagerDownload(path string) (string, []byte) {
	n := filepath.Base(path)
	d, _ := ioutil.ReadFile(filemanagerRootPath + path)
	return n, d
}

func filemanagerResize(path string, width string, height string) []byte {

	resized := models.FromRootDir("cache/preview/" + width + "_" + height + strings.Replace(path, "/", "_", -1))

	if stat, err := os.Stat(resized); err != nil || time.Since(stat.ModTime()) > time.Minute*10 {
		w, _ := strconv.ParseInt(width, 10, 0)
		h, _ := strconv.ParseInt(height, 10, 0)

		f, err := os.Open(models.FromRootDir(filemanagerRootPath + path))
		if err != nil {
			apilog.Println(err)
		}

		img, _, err := image.Decode(f)
		if err != nil {
			apilog.Println(err)
		}

		m := resize.Resize(uint(w), uint(h), img, resize.Lanczos3)

		out, err := os.Create(resized)
		if err != nil {
			apilog.Println(err)
		}
		defer out.Close()

		png.Encode(out, m)
	}

	d, err := ioutil.ReadFile(resized)
	if err != nil {
		apilog.Println(err)
	}

	return d
}

func filemanagerGetInfo(path string) info {

	f, err := os.Lstat(models.FromRootDir(filemanagerRootPath + path))
	if err != nil {
		return info{Error: "Error reading file", Code: 1}
	}

	r := info{
		Path:      path,
		Filename:  f.Name(),
		Protected: 0,
		Properties: properties{
			DateCreated:  "",
			DateModified: string(f.ModTime().String()),
			Height:       "0",
			Width:        "0",
			Size:         strconv.FormatInt(f.Size(), 10),
		},
		Error: "",
		Code:  0,
	}

	ext := filepath.Ext(f.Name())
	if ext != "" {
		ext = ext[1:]
		r.FileType = filepath.Ext(f.Name())[1:]
		if ext == "jpg" || ext == "png" || ext == "gif" {
			r.Preview = "../../filemanager?mode=download&path=" + path + "&width=150&height=0"
		} else {
			ico := "images/fileicons/" + filepath.Ext(f.Name())[1:] + ".png"
			if _, err := os.Stat(ico); err != nil {
				r.Preview = "images/fileicons/default.png"
			} else {
				r.Preview = ico
			}

		}
	}

	return r
}

func filemanagerGetFolder(path string) interface{} {
	r := []info{}
	files, err := ioutil.ReadDir(models.FromRootDir(filemanagerRootPath + path))
	if err != nil {
		return info{
			Error: "Error reading directory",
			Code:  1,
		}
	}

	for _, f := range files {
		t := info{}
		t.Filename = f.Name()
		if f.IsDir() {
			t.Path = path + f.Name() + "/"
			t.FileType = "dir"
			t.Preview = "images/fileicons/_Open.png"
		} else {
			t.Path = path + f.Name()
			ext := filepath.Ext(f.Name())
			if ext != "" {
				ext = ext[1:]
				t.FileType = filepath.Ext(f.Name())[1:]
				if ext == "jpg" || ext == "png" || ext == "gif" {
					t.Preview = "../../filemanager?mode=download&path=" + path + f.Name() + "&width=150&height=0"
				} else {
					ico := "images/fileicons/" + filepath.Ext(f.Name())[1:] + ".png"
					if _, err := os.Stat(ico); err != nil {
						t.Preview = "images/fileicons/default.png"
					} else {
						t.Preview = ico
					}
				}
			}

		}
		r = append(r, t)
	}

	return r

}

func filemanagerDelete(path string) info {

	if err := os.RemoveAll(models.FromRootDir(filemanagerRootPath + path)); err != nil {
		return info{Error: "Error delete", Code: 1}
	}
	return info{Error: "Ok", Code: 0}
}

func filemanagerMkDir(path, name string) info {
	if err := os.MkdirAll(models.FromRootDir(filemanagerRootPath+path+name), 0755); err != nil {
		return info{Error: "Error make directory", Code: 1}
	}
	return info{Error: "Ok", Code: 0}
}

func filemanagerRename(old, new string) info {
	if err := os.Rename(models.FromRootDir(filemanagerRootPath+old), models.FromRootDir(filemanagerRootPath+filepath.Dir(old)+"/"+new)); err != nil {
		return info{Error: "Error rename", Code: 1}
	}
	return info{Error: "Ok", Code: 0}
}
