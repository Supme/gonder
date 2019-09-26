package models

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
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
	logfile, err := NewLogfile(name)
	if err != nil {
		return nil, err
	}
	logger = log.New(io.MultiWriter(logfile, os.Stdout), "", log.Ldate|log.Ltime)
	return logger, nil
}

func NewLogfile(name string) (io.Writer, error) {
	filename := filepath.Join(LogDir, name+".log")
	_, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	logfile := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    0,
		MaxAge:     0,
		MaxBackups: 14,
		LocalTime:  true,
		Compress:   true,
	}
	go func(l *lumberjack.Logger) {
		now := time.Now().Local()
		loc, _ := time.LoadLocation("Local")
		firstTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		if firstTime.Before(now) {
			firstTime = firstTime.Add(time.Hour * 24)
		}
		firstSleep := firstTime.Sub(now)
		time.Sleep(firstSleep)
		for {
			err := l.Rotate()
			if err != nil {
				log.Printf("rotate %s log error: %s\r\n", filepath.Base(l.Filename), err)
			}
			log.Printf("rotate %s log file\r\n", filepath.Base(l.Filename))
			time.Sleep(time.Hour * 24)
		}
	}(logfile)

	return logfile, nil
}
