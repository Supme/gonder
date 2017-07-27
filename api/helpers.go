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
	"log"
	"net/url"
	"regexp"
	"encoding/json"
	"strings"
	"errors"
	"fmt"
	"reflect"
)

func parseArrayForm(r url.Values) map[string]map[string]map[string][]string {
	dblarr := make(map[string]map[string]map[string][]string)
	for key, val := range r {
		res, err := regexp.MatchString(`([\w\d])+(\[([\w\d])+\])(\[([\w\d])+\])`, key)
		if err != nil {
			log.Println(err)
		}
		if res {
			r := regexp.MustCompile(`([\w\d])+`)
			a := r.FindAllString(key, 3)
			if dblarr[a[0]] == nil {
				dblarr[a[0]] = make(map[string]map[string][]string)
			}
			if dblarr[a[0]][a[1]] == nil {
				dblarr[a[0]][a[1]] = make(map[string][]string)
			}
			dblarr[a[0]][a[1]][a[2]] = val
		}
	}
	return dblarr
}

type request struct {
	Cmd string `json:"cmd"`
	Selected []interface{} `json:"selected,omitempty"`
	Limit int64 `json:"limit"`
	Offset int64 `json:"offset"`
	Sort []struct{
		Field string `json:"field"`
		Direction string `json:"direction"`
	} `json:"sort"`
	Search []struct{
		Field string `json:"field"`
		Type string `json:"type"`
		Operator string `json:"operator"`
		Value interface{} `json:"value"`
	} `json:"search,omitempty"`
	SearchLogic string `json:"searchLogic,omitempty"`
	Changes []map[string]interface{} `json:"changes,omitempty"`

	Group int64 `json:"group,omitempty"`
	Campaign int64 `json:"campaign,omitempty"`
	Recipient int64 `json:"recipient,omitempty"`
	FileName string `json:"fileName,omitempty"`
	FileContent string `json:"fileContent,omitempty"`
	Id int64 `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Name string `json:"name,omitempty"`
	Content Data `json:"content,omitempty"`
	Select bool `json:"select,omitempty"`

	DkimSelector  string `json:"dkimSelector"`
	DkimKey  string `json:"dkimKey"`
	DkimUse  int8 `json:"dkimUse"`

	Record struct{
		// Save/Add user
		Id int64 `json:"id,null"`
		Name string `json:"name,omitempty"`
		Password string `json:"password,omitempty"`
		Unit struct{
			Id int64 `json:"id"`
		} `json:"unit,omitempty"`
		Group []struct{
			Id int64 `json:"id"`
		} `json:"group,omitempty"`
		// Unit rights
		GetGroups int8 `json:"get-groups,omitempty"`
		SaveGroups int8 `json:"save-groups,omitempty"`
		AddGroups int8 `json:"add-groups,omitempty"`
		GetCampaigns int8 `json:"get-campaigns,omitempty"`
		SaveCampaigns int8 `json:"save-campaigns,omitempty"`
		AddCampaigns int8 `json:"add-campaigns,omitempty"`
		GetCampaign int8 `json:"get-campaign,omitempty"`
		SaveCampaign int8 `json:"save-campaign,omitempty"`
		GetRecipients int8 `json:"get-recipients,omitempty"`
		GetRecipientParameters int8 `json:"get-recipient-parameters,omitempty"`
		UploadRecipients int8 `json:"upload-recipients,omitempty"`
		DeleteRecipients int8 `json:"delete-recipients,omitempty"`
		GetProfiles int8 `json:"get-profiles,omitempty"`
		AddProfiles int8 `json:"add-profiles,omitempty"`
		DeleteProfiles int8 `json:"delete-profiles,omitempty"`
		SaveProfiles int8 `json:"save-profiles,omitempty"`
		AcceptCampaign int8 `json:"accept-campaign,omitempty"`
		GetLogMain int8 `json:"get-log-main,omitempty"`
		GetLogApi int8 `json:"get-log-api,omitempty"`
		GetLogCampaign int8 `json:"get-log-campaign,omitempty"`
		GetLogUtm int8 `json:"get-log-utm,omitempty"`
	} `json:"record,omitempty"`
}

func parseRequest(js string) (request, error) {
	var req request
	err := json.Unmarshal([]byte(js), &req)
	return req, err
}

func createSqlPart(req request, queryStr string, whereParams []interface{}, mapping map[string]string, withSortLimit bool) (query string, params []interface{}, err error){
	var (
		direction, searchLogic string
		result, srhStr, srtStr []string
	)

	params = whereParams
	if len(req.Search) != 0 {
		result = append(result, "AND")
		if strings.ToUpper(req.SearchLogic) == "OR" {
			searchLogic = " OR "
		} else if strings.ToUpper(req.SearchLogic) == "AND" {
			searchLogic = " AND "
		}
		for _, s := range req.Search {
			if filed, ok := mapping[s.Field]; ok {
				if s.Value != "" {
					var qs string
					if strings.ToLower(s.Type) == "int" {
						if  strings.ToLower(s.Operator) == "more" {
							params = append(params, fmt.Sprintf("%v",s.Value))
							qs = "`" + filed + "`>?"
						} else if  strings.ToLower(s.Operator) == "less" {
							params = append(params, fmt.Sprintf("%v",s.Value))
							qs = "`" + filed + "`<?"
						} else if strings.ToLower(s.Operator) == "between" {
							i := reflect.ValueOf(s.Value).Interface().([]interface{})
							params = append(params, i[0])
							params = append(params, i[1])
							qs = "`" + filed + "` BETWEEN ? AND ?"

						} else {
							params = append(params, fmt.Sprintf("%v",s.Value))
							qs = "`" + filed + "`=?"
						}
					} else if strings.ToLower(s.Type) == "text" {
						if strings.ToLower(s.Operator) == "begins" {
							params = append(params, reflect.ValueOf(s.Value).Interface().(string) + "%")
							qs = "`" + filed + "` LIKE ?"
						} else if strings.ToLower(s.Operator) == "ends" {
							params = append(params, "%" + reflect.ValueOf(s.Value).Interface().(string))
							qs = "`" + filed + "` LIKE ?"
						} else if strings.ToLower(s.Operator) == "contains" {
							params = append(params, "%" + reflect.ValueOf(s.Value).Interface().(string) + "%")
							qs = "`" + filed + "` LIKE ?"
						} else {
							params = append(params, reflect.ValueOf(s.Value).Interface().(string))
							qs = "`" + filed + "`=?"
						}
					}

					srhStr = append(srhStr, qs)
				}
			} else {
				return "", params, errors.New(fmt.Sprintf("field '%s' not in mapping", s.Field))
			}
		}
		if len(srhStr) != 0 {
			result = append(result, " (" + strings.Join(srhStr, searchLogic) + ")")
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
					srtStr = append(srtStr, "`" + filed + "` " + direction)
				} else {
					return "", params, errors.New(fmt.Sprintf("field '%s' not in mapping", s.Field))
				}
			}
			result = append(result, "ORDER BY " + strings.Join(srtStr, ", "))
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

