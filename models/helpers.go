package models

import (
	"encoding/csv"
	"fmt"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Progress struct {
	s map[string]int
	sync.RWMutex
}

func (p *Progress) Set(key string, val int) {
	p.Lock()
	defer p.Unlock()
	if p.s == nil {
		p.s = map[string]int{}
	}
	p.s[key] = val
}

func (p *Progress) Get(key string) *int {
	p.RLock()
	defer p.RUnlock()
	if v, ok := p.s[key]; ok {
		return &v
	}
	return nil
}

func (p *Progress) Delete(key string) {
	p.Lock()
	defer p.Unlock()
	delete(p.s, key)
}

func GetIP(r *http.Request) string {
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		return ipProxy
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func sqlReplaceSpecialSymbols(column string) string {
	return `TRIM(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(` + column + `, "\\", "\\\\"), "\"", "\\\""), "\r", "\\r"), "\n", "\\n"), "\t", "\\t"))`
}

// SQLKeyValueTableToJSON replace JSON_OBJECT for old MySQL/MariaDb
func SQLKeyValueTableToJSON(keyCol, valCol, tableName, where string) string {
	return `CONCAT(
				"{",
				(SELECT GROUP_CONCAT(
    	        	"\"",
					` + sqlReplaceSpecialSymbols(keyCol) + `,
					"\":\"",
					` + sqlReplaceSpecialSymbols(valCol) + `,
					"\""
				ORDER BY ` + keyCol + ` SEPARATOR ",")
				FROM ` + tableName + ` WHERE ` + where + `),
				"}")`
}

func NewCSVWriter(w io.Writer) *csv.Writer {
	var (
		f     io.Writer
		c     *csv.Writer
		comma rune
	)
	switch strings.TrimSpace(strings.ToLower(Config.APIPanelLocale)) {
	case "ru-ru":
		f = transform.NewWriter(w, encoding.ReplaceUnsupported(charmap.Windows1251.NewEncoder()))
		comma = ';'
	default:
		f = w
		comma = ','
	}
	c = csv.NewWriter(f)
	c.Comma = comma
	c.UseCRLF = true
	return c
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
	f, err := os.OpenFile(filename, os.O_CREATE, 0640)
	if err != nil {
		return nil, err
	}
	err = f.Close()
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

// ToDo wrong algoritm for RecipientName <name@domain.tld>
func Bad_GetDomainFromEmail(email string) string {
	if splt := strings.Split(email, "@"); len(splt) == 2 {
		return strings.ToLower(splt[1])
	}
	return ""
}

func GetStatusCodeFromSendResult(res error) string {
	if res == nil {
		return "250"
	}
	code := strings.SplitN(res.Error(), " ", 2)[0]
	_, err := strconv.Atoi(code)
	if err == nil {
		return code
	}
	return res.Error()
}

func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

func StrTimeRFC3339ToReportCSV(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return "error parse time"
	}
	return t.Format(ReportCSVDateFormat)
}
