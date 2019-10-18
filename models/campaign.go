package models

import (
	"database/sql"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/mssola/user_agent"
	"net"
	"strconv"
	"strings"
)

type Campaign int

func (id Campaign) IntID() int {
	return int(id)
}

func (id Campaign) StringID() string {
	return strconv.Itoa(id.IntID())
}

type CampaignReportRecipients struct {
	ID          int             `db:"id" json:"id"`
	Email       string          `db:"email" json:"email"`
	Name        string          `db:"name" json:"name"`
	At          string          `db:"at" json:"at"`
	Status      sql.NullString  `db:"status" json:"-"`
	StatusValid string          `json:"status"`
	Open        bool            `db:"open" json:"open"`
	Data        sql.NullString  `db:"data" json:"-"`
	DataValid   json.RawMessage `json:"data"`
}

func (cq *CampaignReportRecipients) Validate() {
	if cq.Data.Valid {
		cq.DataValid = []byte(cq.Data.String)
	} else {
		cq.DataValid = []byte("null")
	}
	if cq.Status.Valid {
		cq.StatusValid = cq.Status.String
	} else {
		cq.StatusValid = "null"
	}
}

func (id Campaign) ReportRecipients() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT  
  			r.id,
  			r.email AS "email",
  			r.name,
			r.date AS "at",
  			r.status,
  			IF(COALESCE(r.web_agent,r.client_agent) IS NULL, false, true) as "open",
    		`+SQLKeyValueTableToJSON("d.key", "d.value", "parameter d", "d.recipient_id=r.id")+` AS "data"
 		FROM recipient r
 		WHERE r.removed=0 AND r.campaign_id=?`, id)
}

type CampaignReportUnsubscribed struct {
	Email     string          `db:"email" json:"email"`
	At        string          `db:"at" json:"at"`
	Data      sql.NullString  `db:"data" json:"-"`
	DataValid json.RawMessage `json:"data"`
}

func (cq *CampaignReportUnsubscribed) Validate() {
	if cq.Data.Valid {
		cq.DataValid = []byte(cq.Data.String)
	} else {
		cq.DataValid = []byte("null")
	}
}

func (id Campaign) ReportUnsubscribed() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT
        	u.email,
			u.date as "at",
            `+SQLKeyValueTableToJSON("d.name", "d.value", "unsubscribe_extra d", "d.unsubscribe_id=u.id")+` AS "data"
		FROM unsubscribe u
		WHERE u.campaign_id=?`, id)
}

type CampaignReportQuestion struct {
	ID        int             `db:"id" json:"id"`
	Email     string          `db:"email" json:"email"`
	At        string          `db:"at" json:"at"`
	Data      sql.NullString  `db:"data" json:"-"`
	DataValid json.RawMessage `json:"data"`
}

func (cq *CampaignReportQuestion) Validate() {
	if cq.Data.Valid {
		cq.DataValid = []byte(cq.Data.String)
	} else {
		cq.DataValid = []byte("null")
	}
}

func (id Campaign) ReportQuestion() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT
			q.recipient_id AS id,
			r.email,
			q.at,
			`+SQLKeyValueTableToJSON("d.name", "d.value", "question_data d", "d.question_id=q.id")+` AS "data"
		FROM question q
		LEFT JOIN recipient r ON q.recipient_id=r.id
		WHERE r.campaign_id=?`, id)
}

type CampaignReportClicks struct {
	ID    int    `db:"id" json:"id"`
	Email string `db:"email" json:"email"`
	At    string `db:"at" json:"at"`
	URL   string `db:"url" json:"url"`
}

func (cq *CampaignReportClicks) Validate() {}

func (id Campaign) ReportClicks() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT
			r.id,
			r.email,
			j.date as at,
			j.url
		FROM jumping j INNER JOIN recipient r ON j.recipient_id=r.id
		WHERE r.removed=0
			AND j.url NOT IN ('`+OpenTrace+`','`+WebVersion+`','`+Unsubscribe+`')
			AND j.campaign_id=?
		ORDER BY r.id, id`, id)
}

type UserAgent struct {
	IP             string `json:"ip"`
	IsMobile       bool   `json:"is_mobile"`
	IsBot          bool   `json:"is_bot"`
	Platform       string `json:"platform"`
	OS             string `json:"os"`
	EngineName     string `json:"engine_name"`
	EngineVersion  string `json:"engine_version"`
	BrowserName    string `json:"browser_name"`
	BrowserVersion string `json:"browser_version"`
}

func (ua *UserAgent) Parse(str string) {
	split := strings.SplitN(str, " ", 2)
	if len(split) != 2 {
		return
	}
	ua.IP = split[0]
	if ua.isGoogle(ua.IP) {
		ua.IsBot = true
		ua.Platform = "Google bot"
		return
	}
	if ua.isMailRu(ua.IP) {
		ua.IsBot = true
		ua.Platform = "MailRu bot"
		return
	}
	agent := user_agent.New(split[1])
	ua.IsMobile = agent.Mobile()
	ua.IsBot = agent.Bot()
	ua.Platform = agent.Platform()
	ua.OS = agent.OS()
	ua.EngineName, ua.EngineVersion = agent.Engine()
	ua.BrowserName, ua.BrowserVersion = agent.Browser()
}

func (ua UserAgent) isMailRu(ip string) bool {
	// Mail.ru bot from networks
	// 188.93.56.0/24
	// 185.30.176.0/23
	// has user agent
	// Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/32.0.1700.2 Safari/537.36
	return ua.isIPSubnet(ip, "188.93.56.0", 24, 32) || ua.isIPSubnet(ip, "185.30.176.0", 23, 32)
}

func (ua UserAgent) isGoogle(ip string) bool {
	// Google proxy bot from networks
	// 66.102.0.0/20
	// 66.249.64.0/19
	// 64.233.160.0/19
	// has user agent
	// Mozilla/5.0 (Windows NT 5.1; rv:11.0) Gecko Firefox/11.0 (via ggpht.com GoogleImageProxy)
	return ua.isIPSubnet(ip, "66.102.0.0", 20, 32) || ua.isIPSubnet(ip, "66.249.64.0", 19, 32) || ua.isIPSubnet(ip, "64.233.160.0", 19, 32)
}

func (ua UserAgent) isIPSubnet(ip, network string, ones, bits int) bool {
	ipv4Mask := net.CIDRMask(ones, bits)
	return network == net.ParseIP(ip).Mask(ipv4Mask).String()
}

type CampaignReportUserAgent struct {
	ID            int            `db:"id" json:"id"`
	Email         string         `db:"email" json:"email"`
	Name          string         `db:"name" json:"name"`
	Client        sql.NullString `db:"client_agent" json:"-"`
	Browser       sql.NullString `db:"web_agent" json:"-"`
	ClientParsed  *UserAgent     `json:"client"`
	BrowserParsed *UserAgent     `json:"browser"`
}

func (rua *CampaignReportUserAgent) Validate() {
	if rua.Client.Valid {
		rua.ClientParsed = new(UserAgent)
		rua.ClientParsed.Parse(rua.Client.String)
	}
	if rua.Browser.Valid {
		rua.BrowserParsed = new(UserAgent)
		rua.BrowserParsed.Parse(rua.Browser.String)
	}
}

func (id Campaign) ReportUserAgent() (*sqlx.Rows, error) {
	return Db.Queryx(`SELECT id, email, name, client_agent, web_agent FROM recipient WHERE campaign_id=?`, id)
}
