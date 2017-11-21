package models

import (
	"net"
	"net/http"
	"reflect"
	"unsafe"
	"fmt"
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

	if num % 10 == 0 {
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

// Null memory allocate convert
func BytesToString(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	stringHeader := reflect.StringHeader{
		Data: sliceHeader.Data,
		Len:  sliceHeader.Len,
	}
	return *(*string)(unsafe.Pointer(&stringHeader))
}
func StringToBytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sliceHeader := reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
		Cap:  stringHeader.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&sliceHeader))
}
