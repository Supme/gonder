package models

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func GetIP(r *http.Request) string {
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		return ipProxy
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func Conv1st2nd(num int) string {
	var suffix string
	var lastOneNum, lastTwoNum int

	strNum := fmt.Sprintf("%d", num)

	lastOneNum, _ = strconv.Atoi(strNum[len(strNum)-1:])
	if len(strNum) > 2 {
		lastTwoNum, _ = strconv.Atoi(strNum[len(strNum)-2:])
	}

	if num%10 == 0 {
		suffix = "th"
	} else if lastTwoNum >= 11 && lastTwoNum <= 20 {
		suffix = "th"
	} else if lastOneNum == 1 {
		suffix = "st"
	} else if lastOneNum == 2 {
		suffix = "nd"
	} else if lastOneNum == 3 {
		suffix = "rd"
	} else {
		suffix = "th"
	}

	return fmt.Sprintf("%d%s", num, suffix)
}

func NewLogger(name string) (*log.Logger, error) {
	var logger *log.Logger
	l, err := os.OpenFile(filepath.Join(LogDir, name+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return logger, fmt.Errorf("error opening log file: %s", err)
	}
	logger = log.New(io.MultiWriter(l, os.Stdout), "", log.Ldate|log.Ltime)
	return logger, nil
}
