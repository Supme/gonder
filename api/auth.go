package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"gonder/models"
	"log"
	"net/http"
	"net/url"
	"time"
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
		// ToDo rate limit bad auth or/and user requests
		if !authorize {
			//if user != "" {
			//	ip := models.GetIP(r)
			//	apiLog.Printf("%s bad user login '%s'", ip, user)
			//	if models.Config.GonderEmail != "" && models.Config.AdminEmail != "" {
			//		go func() {
			//			bldr := &smtpSender.Builder{
			//				From: models.Config.GonderEmail,
			//				To: models.Config.AdminEmail,
			//				Subject: "Bad login to Gonder",
			//			}
			//			bldr.AddTextPlain([]byte(ip + " bad user login '" + user + "'"))
			//
			//			email := bldr.Email("", func(result smtpSender.Result){
			//				apiLog.Print("Error send mail:", result.Err)
			//			})
			//			email.Send(&smtpSender.Connect{}, nil)
			//		}()
			//	}
			//}
			models.Prometheus.Api.AuthRequest.Inc()
			time.Sleep(time.Second * 1)
			w.Header().Set("WWW-Authenticate", `Basic realm="Gonder"`)
			w.WriteHeader(401)
			return
		}
		auth.name = user

		uri, err := url.QueryUnescape(r.RequestURI)
		if err != nil {
			log.Println(err)
		}
		ip := models.GetIP(r)
		apiLog.Printf("host: %s user: '%s' %s %s", ip, auth.name, r.Method, uri)
		models.Prometheus.Api.UserRequest.WithLabelValues(ip).Inc()
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
	return a.unitID == 0
}

func check(user, password string) (int64, int64, bool) {
	l := false
	var passwordHash string
	var userID, unitID int64

	hash := sha256.New()
	if _, err := hash.Write([]byte(password)); err != nil {
		apiLog.Print(err)
	}
	md := hash.Sum(nil)
	shaPassword := hex.EncodeToString(md)

	err := models.Db.QueryRow("SELECT `id`, `auth_unit_id`, `password` FROM `auth_user` WHERE `name`=?", user).Scan(&userID, &unitID, &passwordHash)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}

	if shaPassword == passwordHash {
		l = true
	}

	return userID, unitID, l
}

// ToDo
func Logout(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Authorization", "Basic")
	http.Error(w, "Logout. Bye!", http.StatusUnauthorized)
}
