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
	"net/http"
	"crypto/sha256"
	"github.com/supme/gonder/models"
	"log"
	"encoding/hex"
	"time"
)

type Auth struct {
	Name string
	userId int64
	unitId int64
}

func (a *Auth) Check(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var authorize bool
		loging(r)

		user, password, _ := r.BasicAuth()
		a.userId, a.unitId, authorize = check(user, password)

		if !authorize {
			w.Header().Set("WWW-Authenticate", `Basic realm="Gonder"`)
			w.WriteHeader(401)
			return
		}

		a.Name = user

		fn(w, r)
	}
}

func (a *Auth) Log(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loging(r)
		fn(w, r)
	}
}

func (a *Auth) GroupRight(group int64) bool {
	var r bool
	var c int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `auth_user_group` WHERE `auth_user_id`=? AND `group_id`=?", a.userId, group).Scan(&c)
	if err != nil {
		r = false
	}

	return r
}

func (a *Auth) Right(right string) bool {
	var r bool

	if a.IsAdmin() {
		log.Println("Admin right")
		return true
	}

	err := models.Db.QueryRow("SELECT COUNT(auth_right.id) user_right FROM `auth_user` JOIN `auth_unit_right` ON auth_user.auth_unit_id = auth_unit_right.auth_unit_id JOIN `auth_right` ON auth_unit_right.auth_right_id = auth_right.id WHERE auth_user.id = ? AND auth_right.name = ?", a.userId, right).Scan(&r)
	if err != nil {
		return false
	}
	log.Println("Right ", r)

	return r
}

func (a *Auth) IsAdmin() bool {
	// admins has group 0
	if a.unitId == 0 {
		return true
	}
	return false
}

func check(user, password string) (int64, int64, bool) {
	l := false
	var passwordHash string
	var userId, unitId int64

	hash := sha256.New()
	hash.Write([]byte(password))
	md := hash.Sum(nil)
	shaPassword := hex.EncodeToString(md)
	//log.Print(string(shaPassword))

	err := models.Db.QueryRow("SELECT `id`, `auth_unit_id`, `password` FROM `auth_user` WHERE `name`=?", user).Scan(&userId, &unitId, &passwordHash)
	if err != nil {
		l = false
	}

	if shaPassword == passwordHash {
		l = true
	}

	return userId, unitId, l
}

//ToDo
func (a *Auth)Logout(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Authorization", "Basic")
	http.Error(w, "Logout. Bye!", 401)
}

func loging(r *http.Request)  {
	log.Printf("%s - %s Host: %s URI: %s Client: %v", time.Now().Format("2006-01-02T15:04:05"), r.Method, r.RemoteAddr, r.RequestURI, r.UserAgent())
}