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
package panel

import "log"

var Port string

type iFace struct {
	Id     string
	Name   string
	Iface  string
	Host   string
	Stream string
	Delay  string
}

type campaign struct {
	Id        string
	IfaceId   string
	Name      string
	Subject   string
	From      string
	FromName  string
	Message   string
	StartTime string
	EndTime   string
	SendUnsubscribe string
}

type recipient struct {
	Id         string
	CampaignId string
	Email      string
	Name       string
}

func Run()  {
	routers()
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
