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
	"github.com/hpcloud/tail"
	"golang.org/x/net/websocket"
	"os"
)

func CampaignLog(ws *websocket.Conn) {
	statusLog(ws, "./log/campaign.log")
}

func ApiLog(ws *websocket.Conn) {
	statusLog(ws, "./log/api.log")
}

func StatisticLog(ws *websocket.Conn) {
	statusLog(ws, "./log/statistic.log")
}

func MainLog(ws *websocket.Conn) {
	statusLog(ws, "./log/main.log")
}

func statusLog(ws *websocket.Conn, file string) {
	var err error
	var offset tail.SeekInfo
	offset.Whence = 2
	fi, err := os.Open(file)
	if err != nil {
		apilog.Println(err)
	}
	f, err := fi.Stat()
	if err != nil {
		apilog.Println(err)
	}
	if f.Size() < 2000 {
		offset.Offset = f.Size() * (-1)
	} else {
		offset.Offset = -2000
	}
	fi.Close()

	conf := tail.Config{
		Follow: true,
		ReOpen: true,
		Location: &offset,
		Logger: tail.DiscardingLogger,
	}
	t, err := tail.TailFile(file, conf)
	if err != nil {
		apilog.Println(err)
	}

	go func() {
		if err = websocket.Message.Send(ws, "..."); err != nil {
			apilog.Println("Can't send websocket message", err)
		}
		for line := range t.Lines {
			if err = websocket.Message.Send(ws, line.Text); err != nil {
				apilog.Println("Can't send websocket message", err)
				break
			}
		}
	}()

	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			apilog.Println("Can't receive")
			break
		}

		if reply == "ping" {
			if err = websocket.Message.Send(ws, "pong"); err != nil {
				apilog.Println("Can't send")
				break
			}
		}

	}
}

/*
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("Can't receive")
			break
		}

		fmt.Println("Received back from client: " + reply)

		msg := "Received:  " + reply
		fmt.Println("Sending to client: " + msg)

		if err = websocket.Message.Send(ws, msg); err != nil {
			fmt.Println("Can't send")
			break
		}

	}

 */
