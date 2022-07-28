package models

var (
	AppVersion = "na"
	AppCommit  = "na"
	AppDate    = "na"
)

const (
	OpenTrace       = "open_trace"
	WebVersion      = "web_version"
	UnsubscribeForm = "unsubscribe_form"
	Unsubscribe     = "unsubscribe"

	StatusUnavailableRecentTime = "Unavailable recent time"
	StatusUnsubscribe           = "Unsubscribe"
	StatusSending               = "Sending"
	APILog                      = "gonder_api"
	CampaignLog                 = "gonder_campaign"
	UTMLog                      = "gonder_utm"
	MainLog                     = "gonder_main"

	StatHTMLImgTag = `<img src="{{.StatUrl}}" border="0" width="10" height="10" alt=""/>`
	StatAMPImgTag  = `<amp-img src="{{.AmpStatUrl}}" width="10" height="10" layout="fixed"></amp-img>`

	ReportCSVDateFormat = "2006-01-02 15:04:05"
)
