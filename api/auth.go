package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/supme/directEmail"
	"gonder/models"
	"log"
	"net/http"
	"net/url"
)

type Auth struct {
	name   string
	userID int64
	unitID int64
}

func CheckAuth(fn http.HandlerFunc) http.HandlerFunc {
	auth := new(Auth)
	return func(w http.ResponseWriter, r *http.Request) {
		var authorize bool
		user, password, _ := r.BasicAuth()
		auth.userID, auth.unitID, authorize = check(user, password)
		if !authorize {
			if user != "" {
				ip := models.GetIP(r)
				apilog.Printf("%s bad user login '%s'", ip, user)
				if models.Config.GonderMail != "" && models.Config.AdminMail != "" {
					// ToDo use smtpSender
					go func() {
						email := directEmail.New()
						email.FromEmail = models.Config.GonderMail
						email.ToName = models.Config.AdminMail
						email.Subject = "Bad login to Gonder"
						email.TextPlain(ip + " bad user login '" + user + "'")
						email.Render()
						if err := email.Send(); err != nil {
							apilog.Print("Error send mail:", err)
						}
					}()
				}
			}
			w.Header().Set("WWW-Authenticate", `Basic realm="Gonder"`)
			w.WriteHeader(401)
			return
		}
		auth.name = user

		uri, err := url.QueryUnescape(r.RequestURI)
		if err != nil {
			log.Println(err)
		}
		apilog.Printf("host: %s user: '%s' %s %s", models.GetIP(r), auth.name, r.Method, uri)

		ctx := context.WithValue(r.Context(), "Auth", auth)
		fn(w, r.WithContext(ctx))
	}
}

func (a *Auth) GroupRight(group interface{}) bool {
	if a.IsAdmin() {
		return true
	}

	var r = true
	var c int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `auth_user_group` WHERE `auth_user_id`=? AND `group_id`=?", a.userID, group).Scan(&c)
	if err != nil {
		log.Println(err)
		r = false
	}
	if c == 0 {
		r = false
	}
	return r
}

func (a *Auth) CampaignRight(campaign interface{}) bool {
	if a.IsAdmin() {
		return true
	}

	var r = true
	var c int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `auth_user_group` WHERE `auth_user_id`=? AND `group_id`=(SELECT `group_id` FROM `campaign` WHERE id=?)", a.userID, campaign).Scan(&c)
	if err != nil {
		log.Println(err)
		r = false
	}
	if c == 0 {
		r = false
	}
	return r
}

func (a *Auth) Right(right string) bool {
	var r bool

	if a.IsAdmin() {
		return true
	}

	err := models.Db.QueryRow("SELECT COUNT(auth_right.id) user_right FROM `auth_user` JOIN `auth_unit_right` ON auth_user.auth_unit_id = auth_unit_right.auth_unit_id JOIN `auth_right` ON auth_unit_right.auth_right_id = auth_right.id WHERE auth_user.id = ? AND auth_right.name = ?", a.userID, right).Scan(&r)
	if err != nil {
		log.Println(err)
		return false
	}

	return r
}

func (a *Auth) IsAdmin() bool {
	// admins has group 0
	if a.unitID == 0 {
		return true
	}
	return false
}

func check(user, password string) (int64, int64, bool) {
	l := false
	var passwordHash string
	var userID, unitID int64

	hash := sha256.New()
	hash.Write([]byte(password))
	md := hash.Sum(nil)
	shaPassword := hex.EncodeToString(md)

	err := models.Db.QueryRow("SELECT `id`, `auth_unit_id`, `password` FROM `auth_user` WHERE `name`=?", user).Scan(&userID, &unitID, &passwordHash)
	if err != nil {
		log.Println(err)
		l = false
	}

	if shaPassword == passwordHash {
		l = true
	}

	return userID, unitID, l
}

//ToDo
func Logout(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Authorization", "Basic")
	http.Error(w, "Logout. Bye!", 401)
}
