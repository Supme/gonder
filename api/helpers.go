package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gonder/models"
	"log"
	"reflect"
	"strings"
)

type campaignData struct {
	ID              int64  `json:"recid"`
	Name            string `json:"name"`
	ProfileID       int    `json:"profileId"`
	Subject         string `json:"subject"`
	SenderID        int    `json:"senderId"`
	StartDate       int64  `json:"startDate"`
	EndDate         int64  `json:"endDate"`
	CompressHTML    bool   `json:"compressHTML"`
	SendUnsubscribe bool   `json:"sendUnsubscribe"`
	Accepted        bool   `json:"accepted"`
	TemplateHTML    string `json:"templateHTML"`
	TemplateText    string `json:"templateText"`
	TemplateAMP     string `json:"templateAMP"`
}

type request struct {
	auth *Auth

	Cmd      string        `json:"cmd"`
	Selected []interface{} `json:"selected,omitempty"`
	Limit    int64         `json:"limit"`
	Offset   int64         `json:"offset"`
	Sort     []struct {
		Field     string `json:"field"`
		Direction string `json:"direction"`
	} `json:"sort"`
	Search []struct {
		Field    string      `json:"field"`
		Type     string      `json:"type"`
		Operator string      `json:"operator"`
		Value    interface{} `json:"value"`
	} `json:"search,omitempty"`
	SearchLogic string                   `json:"searchLogic,omitempty"`
	Changes     []map[string]interface{} `json:"changes,omitempty"`

	Group       int64        `json:"group,omitempty"`
	Campaign    int64        `json:"campaign,omitempty"`
	Recipient   int64        `json:"recipient,omitempty"`
	FileName    string       `json:"fileName,omitempty"`
	FileContent []byte       `json:"fileContent,omitempty"`
	ID          int64        `json:"id,omitempty"`
	Email       string       `json:"email,omitempty"`
	Name        string       `json:"name,omitempty"`
	UtmURL      string       `json:"utmURL,omitempty"`
	Content     campaignData `json:"content,omitempty"`
	Select      bool         `json:"select,omitempty"`

	Recipients recips `json:"recipients,omitempty"`
	IDs        []int  `json:"ids,omitempty"`

	BimiSelector string `json:"bimiSelector"`

	DkimSelector string `json:"dkimSelector"`
	DkimKey      string `json:"dkimKey"`
	DkimUse      int8   `json:"dkimUse"`

	Interval int `json:"interval"`

	Record struct {
		// Save/Add user
		ID              int64  `json:"id,omitempty"`
		Name            string `json:"name,omitempty"`
		Password        string `json:"password,omitempty"`
		NewPassword     string `json:"newPassword,omitempty"`
		ConfirmPassword string `json:"confirmPassword,omitempty"`
		Blocked         int8   `json:"blocked,omitempty"`
		Unit            struct {
			ID int64 `json:"id"`
		} `json:"unit,omitempty"`
		Group []struct {
			ID int64 `json:"id"`
		} `json:"group,omitempty"`
		// Unit rights
		GetGroups              int8 `json:"get-groups,omitempty"`
		SaveGroups             int8 `json:"save-groups,omitempty"`
		AddGroups              int8 `json:"add-groups,omitempty"`
		GetCampaigns           int8 `json:"get-campaigns,omitempty"`
		SaveCampaigns          int8 `json:"save-campaigns,omitempty"`
		AddCampaigns           int8 `json:"add-campaigns,omitempty"`
		GetCampaign            int8 `json:"get-campaign,omitempty"`
		SaveCampaign           int8 `json:"save-campaign,omitempty"`
		GetRecipients          int8 `json:"get-recipients,omitempty"`
		GetRecipientParameters int8 `json:"get-recipient-parameters,omitempty"`
		UploadRecipients       int8 `json:"upload-recipients,omitempty"`
		DeleteRecipients       int8 `json:"delete-recipients,omitempty"`
		GetProfiles            int8 `json:"get-profiles,omitempty"`
		AddProfiles            int8 `json:"add-profiles,omitempty"`
		DeleteProfiles         int8 `json:"delete-profiles,omitempty"`
		SaveProfiles           int8 `json:"save-profiles,omitempty"`
		AcceptCampaign         int8 `json:"accept-campaign,omitempty"`
		GetLogMain             int8 `json:"get-log-main,omitempty"`
		GetLogAPI              int8 `json:"get-log-api,omitempty"`
		GetLogCampaign         int8 `json:"get-log-campaign,omitempty"`
		GetLogUtm              int8 `json:"get-log-utm,omitempty"`
	} `json:"record,omitempty"`
}

func parseRequest(js []byte) (request, error) {
	var req request
	err := json.Unmarshal(js, &req)
	if err != nil {
		apiLog.Print(err)
		err = fmt.Errorf("parse request unmarshal %s", err)
	}
	return req, err
}

func createSQLPart(req request, queryStr string, whereParams []interface{}, mapping map[string]string, withSortLimit bool) (query string, params []interface{}, err error) {
	var (
		direction, searchLogic string
		result, srhStr, srtStr []string
	)

	params = whereParams
	if len(req.Search) != 0 {
		result = append(result, "AND")
		switch strings.ToUpper(req.SearchLogic) {
		case "OR":
			searchLogic = " OR "
		case "AND":
			searchLogic = " AND "
		default:
			searchLogic = " OR "
		}
		for _, s := range req.Search {
			if filed, ok := mapping[s.Field]; ok {
				if s.Value != "" {
					var qs string
					switch strings.ToLower(s.Type) {
					case "int":
						switch strings.ToLower(s.Operator) {
						case "more", ">":
							params = append(params, fmt.Sprintf("%v", s.Value))
							qs = "`" + filed + "`>?"
						case ">=":
							params = append(params, fmt.Sprintf("%v", s.Value))
							qs = "`" + filed + "`>=?"
						case "less", "<":
							params = append(params, fmt.Sprintf("%v", s.Value))
							qs = "`" + filed + "`<?"
						case "<=":
							params = append(params, fmt.Sprintf("%v", s.Value))
							qs = "`" + filed + "`<=?"
						case "between":
							i := reflect.ValueOf(s.Value).Interface().([]interface{})
							params = append(params, i[0])
							params = append(params, i[1])
							qs = "`" + filed + "` BETWEEN ? AND ?"
						default:
							params = append(params, fmt.Sprintf("%v", s.Value))
							qs = "`" + filed + "`=?"
						}
					case "text":
						switch strings.ToLower(s.Operator) {
						case "begins":
							params = append(params, reflect.ValueOf(s.Value).Interface().(string)+"%")
							qs = "`" + filed + "` LIKE ?"
						case "ends":
							params = append(params, "%"+reflect.ValueOf(s.Value).Interface().(string))
							qs = "`" + filed + "` LIKE ?"
						case "contains":
							params = append(params, "%"+reflect.ValueOf(s.Value).Interface().(string)+"%")
							qs = "`" + filed + "` LIKE ?"
						default:
							params = append(params, reflect.ValueOf(s.Value).Interface().(string))
							qs = "`" + filed + "`=?"
						}
					}
					srhStr = append(srhStr, qs)
				}
			} else {
				return "", params, fmt.Errorf("field '%s' not in mapping", s.Field)
			}
		}
		if len(srhStr) != 0 {
			result = append(result, " ("+strings.Join(srhStr, searchLogic)+")")
		} else {
			result = append(result, "1=1")
		}
	}
	if withSortLimit {
		if len(req.Sort) != 0 {
			for _, s := range req.Sort {
				if strings.ToUpper(s.Direction) == "ASC" {
					direction = "ASC"
				} else if strings.ToUpper(s.Direction) == "DESC" {
					direction = "DESC"
				}
				if filed, ok := mapping[s.Field]; ok {
					srtStr = append(srtStr, "`"+filed+"` "+direction)
				} else {
					return "", params, fmt.Errorf("field '%s' not in mapping", s.Field)
				}
			}
			result = append(result, "ORDER BY "+strings.Join(srtStr, ", "))
		}

		if req.Limit != 0 {
			result = append(result, fmt.Sprintf("LIMIT %d", req.Limit))
		}

		if req.Limit != 0 && req.Offset != 0 {
			result = append(result, fmt.Sprintf("OFFSET %d", req.Offset))
		}
	}

	query = queryStr + " " + strings.Join(result, " ")
	return query, params, nil
}

func isAccepted(campaignID int64) bool {
	var accepted bool
	err := models.Db.QueryRow("SELECT `accepted` FROM campaign WHERE id=?", campaignID).Scan(&accepted)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return accepted
}
