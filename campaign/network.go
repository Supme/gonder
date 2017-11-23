package campaign

import (
	"github.com/supme/gonder/models"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type mxStor struct {
	records []*net.MX
	update  time.Time
}

var mx = struct {
	stor map[string]mxStor
	sync.Mutex
}{
	stor: make(map[string]mxStor),
}

// domainGetMX resolv domain MX records
func domainGetMX(domain string) ([]*net.MX, error) {
	var (
		record []*net.MX
		err    error
	)

	if models.Config.DNScache {
		mx.Lock()
		defer mx.Unlock()
		if _, ok := mx.stor[domain]; !ok || time.Since(mx.stor[domain].update) > 15*time.Minute {
			record, err = net.LookupMX(domain)
			if err == nil {
				mx.stor[domain] = mxStor{
					records: record,
					update:  time.Now(),
				}
			} else if _, ok := mx.stor[domain]; ok {
				record = mx.stor[domain].records
			}
		} else {
			record = mx.stor[domain].records
		}
	} else {
		record, err = net.LookupMX(domain)
	}

	return record, err
}

type (
	profileData struct {
		iface, host          string
		streamNow, streamMax int
		lastUpdate           time.Time
	}
)

var (
	profileStor  = map[int]profileData{}
	profileGroup = map[int]int{}
	profileMutex sync.Mutex
)

// profileNext get next free sending profile data and add connection count
func profileNext(id int) (int, string, string) {
	var res profileData

	profileMutex.Lock()

	// Если есть в массиве и недавно обновлялось
	if _, ok := profileStor[id]; ok && time.Since(profileStor[id].lastUpdate) < 60*time.Second {

		// Если этот профиль группа
		if strings.ToLower(strings.TrimSpace(profileStor[id].host)) == "group" {
			if _, gok := profileGroup[id]; !gok {
				profileGroup[id] = 0
			}
			gIfaces := strings.Split(profileStor[id].iface, ",")
			if profileGroup[id]+1 > len(gIfaces) {
				profileGroup[id] = 0
			}
			i, e := strconv.Atoi(strings.TrimSpace(gIfaces[profileGroup[id]]))
			if e != nil {
				log.Print(e)
			}
			profileGroup[id]++

			profileMutex.Unlock()
			return profileNext(i)
		}

		// Не достигли максимума потоков
		if profileStor[id].streamNow < profileStor[id].streamMax {
			res = profileStor[id]
			res.streamNow++
			profileStor[id] = res
			profileMutex.Unlock()
			return id, res.iface, res.host
		}
		// достигли максимума потоков, ждём освобождения
		profileMutex.Unlock()
		for !profileCheck(id) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
		}
		return profileNext(id)

	}

	// В остальных случаях обновляем данные
	err := models.Db.QueryRow("SELECT `iface`,`host`,`stream` FROM `profile` WHERE `id`=?", id).Scan(&res.iface, &res.host, &res.streamMax)
	if err != nil {
		log.Print(err)
	}
	// если уже существовало, сохраним
	if _, ok := profileStor[id]; ok {
		res.streamNow = profileStor[id].streamNow
	}
	res.lastUpdate = time.Now()
	profileStor[id] = res

	// и повторяем действие
	profileMutex.Unlock()
	return profileNext(id)

}

// profileCheck check profile is busy (max coonnection)?
func profileCheck(id int) bool {
	var free bool
	profileMutex.Lock()
	free = profileStor[id].streamNow < profileStor[id].streamMax
	profileMutex.Unlock()
	return free
}

// profileFree free one connection profile
func profileFree(id int) {
	var res profileData

	profileMutex.Lock()
	res = profileStor[id]
	res.streamNow--
	profileStor[id] = res
	profileMutex.Unlock()
	log.Println("profile id =", id, " connection count =", res.streamNow+1)
}
