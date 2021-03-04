package api

import (
	"context"
	"gonder/models"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Auth struct {
	user *models.User
}

func AuthHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		auth := new(Auth)
		username, password, _ := r.BasicAuth()
		auth.user, err = models.UserGetByNameAndPassword(username, password)
		if err != nil {
			if err != models.ErrUserNotFound || err != models.ErrUserInvalidPassword {
				log.Print(err)
			}
			// ToDo rate limit bad auth or/and user requests
			models.Prometheus.Api.AuthRequest.Inc()
			time.Sleep(time.Second * 1)
			w.Header().Set("WWW-Authenticate", `Basic realm="Gonder"`)
			w.WriteHeader(401)
			return
		}

		uri, err := url.QueryUnescape(r.RequestURI)
		if err != nil {
			log.Println(err)
		}

		ip := models.GetIP(r)
		apiLog.Printf("host: %s user: '%s' %s %s", ip, auth.user.Name, r.Method, uri)
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
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `auth_user_group` WHERE `auth_user_id`=? AND `group_id`=?", a.user.ID, group).Scan(&c)
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
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `auth_user_group` WHERE `auth_user_id`=? AND `group_id`=(SELECT `group_id` FROM `campaign` WHERE id=?)", a.user.ID, campaign).Scan(&c)
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

	err := models.Db.QueryRow("SELECT COUNT(auth_right.id) user_right FROM `auth_user` JOIN `auth_unit_right` ON auth_user.auth_unit_id = auth_unit_right.auth_unit_id JOIN `auth_right` ON auth_unit_right.auth_right_id = auth_right.id WHERE auth_user.id = ? AND auth_right.name = ?", a.user.ID, right).Scan(&r)
	if err != nil {
		log.Println(err)
		return false
	}

	return r
}

func (a *Auth) IsAdmin() bool {
	return a.user.IsAdmin()
}

// ToDo
func Logout(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Authorization", "Basic")
	http.Error(w, "Logout. Bye!", http.StatusUnauthorized)
}
