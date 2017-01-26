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
		Operator string `json:"operator"`
		Value string `json:"value"`
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
	Record struct{
		// Save/Add user
		Id int64 `json:"id,null"`
		Name string `json:"name,omitempty"`
		Password string `json:"password,omitempty"`
		Unit struct{
			Id int64 `json:"id"`
		} `json:"unit,omitempty"`
		Group []struct{
			Id int64 `json:"id,omitempty,-"`
		} `json:"group,omitempty"`
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
					srhStr = append(srhStr, "`" + filed + "` LIKE ?")
					if strings.ToLower(s.Operator) == "begins" {
						params = append(params, s.Value + "%")
					} else if strings.ToLower(s.Operator) == "ends" {
						params = append(params, "%" + s.Value)
					} else if strings.ToLower(s.Operator) == "contains" {
						params = append(params, "%" + s.Value + "%")
					} else {
						params = append(params, s.Value)
					}
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