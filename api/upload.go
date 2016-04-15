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
	"fmt"
	"net/http"
	"log"
	"encoding/base64"
	"io/ioutil"
	"time"
	"math/rand"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	log.Print(r.FormValue("name"))
	log.Print(r.FormValue("type"))
	log.Print(r.FormValue("size"))
	log.Print(r.FormValue("content"))

	file, err := base64.StdEncoding.DecodeString(r.FormValue("content"))
	err = ioutil.WriteFile("./tmp/" + r.FormValue("name"), file, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, 	r.FormValue("name"))

	wait := time.Duration(rand.Int()/10000000000) * time.Nanosecond
	time.Sleep(wait)

}

/*

------WebKitFormBoundaryYs3c1v4vZcIxwog0
Content-Disposition: form-data; name="file"; filename="_gosender.csv"
Content-Type: application/vnd.ms-excel


------WebKitFormBoundaryYs3c1v4vZcIxwog0
Content-Disposition: form-data; name="submit"

Submit
------WebKitFormBoundaryYs3c1v4vZcIxwog0--

 */



/*

------WebKitFormBoundarydthJITFFjCPn8Wws
Content-Disposition: form-data; name="file"

ZW1haWwsbmFtZSxOYW1lLEFnZSxHZW5kZXINCmEuYWdhZm9ub3ZAZG1iYXNpcy5ydSxMeW9zaGEsQWdhZm9ub3YgQWxleGV5LDM2LG0NCm4ua296a2luQGRtYmFzaXMucnUsS29seWEsS296a2luIE5pa29sYXksMjcsbQ0KcC50eXVtZW5ldkBkbWJhc2lzLnJ1LFBhc2hhLFR5dW1lbmV2IFBhdmVsLDM2LG0NCm5pa29sYXlAa296a2luLnJ1LEtvbHlhLEtvemtpbiBOaWtvbGF5LDI3LG0NCnN1cG1lYUBnbWFpbC5jb20sU3VwbWUsLCwNCnN1cG1lQG5ncy5ydSzQkNC70LXQutGB0LXQuSwsLG0NCnMuc2l0ZGlrb3ZAZG1iYXNpcy5ydSzQodC10YDQs9C10Lks0KHQtdGA0LPQtdC5INCh0LjRgtC00LjQutC+0LIs0L3QtdGB0LrQvtC70YzQutC+INC70LXRgizQvNGD0LbRgdC60L7QuQ0K
------WebKitFormBoundarydthJITFFjCPn8Wws--

 */