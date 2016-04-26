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
package main

import (
	"github.com/supme/gonder/statistic"
//	"github.com/supme/gonder/panel"
	"github.com/supme/gonder/models"
	"github.com/supme/gonder/campaign"
	"fmt"
	"github.com/alyu/configparser"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"os/exec"
	"runtime"
	"database/sql"
	"io"
	"bufio"
	"syscall"
	"strconv"
	"errors"
	"github.com/supme/gonder/api"
)

func main() {

	l, err := os.OpenFile("log/main.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Println("error opening log file: %v", err)
	}
	defer l.Close()

	ml := io.MultiWriter(l, os.Stdout)

	log.SetFlags(3)
	log.SetOutput(ml)


	runtime.GOMAXPROCS(runtime.NumCPU())

	config, err := configparser.Read("./config.ini")
	checkErr(err)

	mainConfig, err := config.Section("Main")
	checkErr(err)

	dbConfig, err := config.Section("Database")
	checkErr(err)

	mailerConfig, err := config.Section("Mailer")
	checkErr(err)

	statisticConfig, err := config.Section("Statistic")
	checkErr(err)

	apiConfig, err := config.Section("API")
	checkErr(err)

	// Init models
	models.Db, err = sql.Open(dbConfig.ValueOf("type"), dbConfig.ValueOf("string"))
	checkErr(err)
	defer models.Db.Close()
	checkErr(models.Db.Ping())

	models.Db.SetMaxIdleConns(10)
	models.Db.SetMaxOpenConns(10)

	models.Version = "Gonder 0.6"
	models.StatUrl = "http://" + mainConfig.ValueOf("host")

	// Init mailer
	if mailerConfig.ValueOf("send") == "yes" {
		campaign.Send = true
	} else {
		campaign.Send = false
	}
	campaign.MaxCampaingns, err = strconv.Atoi(mailerConfig.ValueOf("maxcampaign"))
	if err != nil {
		panic("Error parse config parametr 'maxcampaign'")
	}

	// Init statistic
	statistic.Port = statisticConfig.ValueOf("port")

	// Init api
	api.Port = apiConfig.ValueOf("port")

	// Start
	if len(os.Args) == 2 {
		var err error
		if os.Args[1] == "status" {
			err = checkPid("api")
			if err == nil {
				fmt.Println("Process api running")
			} else {
				fmt.Println("Process api stoping")
			}
			err = checkPid("sender")
			if err == nil {
				fmt.Println("Process sender running")
			} else {
				fmt.Println("Process sender stoping")
			}
			err = checkPid("stat")
			if err == nil {
				fmt.Println("Process stat running")
			} else {
				fmt.Println("Process stat stoping")
			}
		}
		if os.Args[1] == "start" {
			err = startProcess("api")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = startProcess("sender")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = startProcess("stat")
			if err != nil {
				fmt.Println(err.Error())
			}
			os.Exit(0)
		}
		if os.Args[1] == "stop" {
			err = stopProcess("api")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = stopProcess("sender")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = stopProcess("stat")
			if err != nil {
				fmt.Println(err.Error())
			}
			os.Exit(0)
		}
		if os.Args[1] == "restart" {
			err = stopProcess("api")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = startProcess("api")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = stopProcess("stat")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = startProcess("stat")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = stopProcess("sender")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = startProcess("sender")
			if err != nil {
				fmt.Println(err.Error())
			}
			os.Exit(0)
		}
	}

	if len(os.Args) == 3 {
		if os.Args[1] == "start" {
			if os.Args[2] == "api" {
				err = startProcess("api")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			if os.Args[2] == "sender" {
				err = startProcess("sender")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			if os.Args[2] == "stat" {
				err = startProcess("stat")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			os.Exit(0)
		}

		if os.Args[1] == "stop" {
			if os.Args[2] == "api" {
				err = stopProcess("api")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			if os.Args[2] == "sender" {
				err = stopProcess("sender")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			if os.Args[2] == "stat" {
				err = stopProcess("stat")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			os.Exit(0)
		}

		if os.Args[1] == "restart" {
			if os.Args[2] == "api" {
				err = stopProcess("api")
				if err != nil {
					fmt.Println(err.Error())
				}
				err = startProcess("api")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			if os.Args[2] == "sender" {
				err = stopProcess("sender")
				if err != nil {
					fmt.Println(err.Error())
				}
				err = startProcess("sender")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			if os.Args[2] == "stat" {
				err = stopProcess("stat")
				if err != nil {
					fmt.Println(err.Error())
				}
				err = startProcess("stat")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			os.Exit(0)
		}

		if os.Args[1] == "daemonize" {

			if os.Args[2] == "api" {

				fmt.Println("Start api http server")
				api.Run()

				for {}
			}

			if os.Args[2] == "sender" {

				fmt.Println("Start database mailer")
				campaign.Run()

				for {}
			}

			if os.Args[2] == "stat" {

				fmt.Println("Start statistics http server")
				statistic.Run()

				for {}
			}

		}

	}

	if len(os.Args) == 1 {
		fmt.Println("Start api http server")
		go api.Run()

		fmt.Println("Start database mailer")
		go campaign.Run()

		fmt.Println("Start statistics http server")
		go statistic.Run()

		fmt.Println("Press Enter for stop")
		var input string
		fmt.Scanln(&input)
	}


}

func startProcess(name string) error {
	err := checkPid(name)
	if err == nil {
		return errors.New("Process " + name + " already running")
	} else {
		p := exec.Command(os.Args[0], "daemonize", name)
		p.Start()
		fmt.Println("Started " + name + " pid", p.Process.Pid)
		err := setPid(name, p.Process.Pid)
		if err != nil {
			return errors.New(name + " set PID error: " + err.Error())
		}
	}

	return  nil
}

func stopProcess(name string) error {
	err := checkPid(name)
	if err != nil {
		fmt.Println("Process " + name + " not found:")
		return err
	} else {
		file, err := os.Open("pid/" + name + ".pid")
		if err != nil {
			return err
		}
		reader := bufio.NewReader(file)
		pid, _, err :=reader.ReadLine()
		if err != nil {
			return err
		}
		p, _ := strconv.Atoi(string(pid))
		process, _ := os.FindProcess(p)
		err = process.Kill()
		if err != nil {
			return err
		}
		os.Remove("pid/" + name + ".pid")
	}
	fmt.Println("Process " + name + " stoped")
	return nil
}

func setPid(name string, pid int) error {
	p := strconv.Itoa(pid)
	file, err := os.Create("pid/" + name + ".pid")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.WriteString(file, p)
	if err != nil {
		return err
	}
	return nil
}

func checkPid(name string) error {
	file, err := os.Open("pid/" + name + ".pid")
	if err != nil {
		return err
	}
	reader := bufio.NewReader(file)
	pid, _, err :=reader.ReadLine()
	if err != nil {
		return err
	}

	p, _ := strconv.Atoi(string(pid))
	process, err := os.FindProcess(p)
	if err != nil {
		os.Remove("pid/" + name + ".pid")
		return errors.New("Failed to find process")
	} else {
		err := process.Signal(syscall.Signal(0))
		if err != nil {
			os.Remove("pid/" + name + ".pid")
			return errors.New("Process not response to signal.")
		}
	}

	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
