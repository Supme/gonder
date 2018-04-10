package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/supme/gonder/api"
	"github.com/supme/gonder/campaign"
	"github.com/supme/gonder/models"
	"github.com/supme/gonder/utm"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

var (
	errFailedFindProcess  = errors.New("Failed to find process")
	errProcessNotResponse = errors.New("Process not response to signal.")
)

// Run starting gonder from command line
func Run() {
	var err error
	if len(os.Args) == 2 {
		if os.Args[1] == "status" {
			err = checkPid("sender")
			if err == nil {
				fmt.Println("Process sender running")
			} else {
				fmt.Println("Process sender stoping")
			}
			err = checkPid("utm")
			if err == nil {
				fmt.Println("Process utm running")
			} else {
				fmt.Println("Process utm stoping")
			}
		}
		if os.Args[1] == "start" {
			err = startProcess("sender")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = startProcess("utm")
			if err != nil {
				fmt.Println(err.Error())
			}
			os.Exit(0)
		}
		if os.Args[1] == "stop" {
			err = stopProcess("sender")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = stopProcess("utm")
			if err != nil {
				fmt.Println(err.Error())
			}
			os.Exit(0)
		}
		if os.Args[1] == "restart" {
			err = stopProcess("utm")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = startProcess("utm")
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
			if os.Args[2] == "sender" {
				err = startProcess("sender")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			if os.Args[2] == "utm" {
				err = startProcess("utm")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			os.Exit(0)
		}

		if os.Args[1] == "stop" {
			if os.Args[2] == "sender" {
				err = stopProcess("sender")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			if os.Args[2] == "utm" {
				err = stopProcess("utm")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			os.Exit(0)
		}

		if os.Args[1] == "restart" {
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
			if os.Args[2] == "utm" {
				err = stopProcess("utm")
				if err != nil {
					fmt.Println(err.Error())
				}
				err = startProcess("utm")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			os.Exit(0)
		}

		if os.Args[1] == "daemonize" {
			models.Init()
			if os.Args[2] == "sender" {
				fmt.Println("Start campaign mailer")
				go campaign.Run()
				fmt.Println("Start api http server")
				api.Run()
			}

			if os.Args[2] == "utm" {
				fmt.Println("Start utm http server")
				utm.Run()
			}

		}

	}

	if len(os.Args) == 1 {
		models.Init()
		fmt.Println("Start api http server")
		go api.Run()

		fmt.Println("Start database mailer")
		go campaign.Run()

		fmt.Println("Start utm http server")
		go utm.Run()

		fmt.Println("Press Enter for stop")
		var input string
		fmt.Scanln(&input)
	}

}

func startProcess(name string) error {
	err := checkPid(name)
	if err == nil {
		return errors.New("Process " + name + " already running")
	}
	p := exec.Command(os.Args[0], "daemonize", name)
	p.Start()
	fmt.Println("Started "+name+" pid", p.Process.Pid)
	err = setPid(name, p.Process.Pid)
	if err != nil {
		return errors.New(name + " set PID error: " + err.Error())
	}

	return nil
}

func stopProcess(name string) error {
	err := checkPid(name)

	file, err := os.Open(models.FromRootDir("pid/" + name + ".pid"))
	if err != nil {
		return err
	}
	reader := bufio.NewReader(file)
	pid, _, err := reader.ReadLine()
	if err != nil {
		return err
	}
	p, _ := strconv.Atoi(string(pid))
	process, _ := os.FindProcess(p)
	if err != nil {
		os.Remove(models.FromRootDir("pid/" + name + ".pid"))
		return errFailedFindProcess
	}
	// ToDo wait
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	os.Remove(models.FromRootDir("pid/" + name + ".pid"))

	fmt.Println("Process " + name + " stoped")
	return nil
}

func setPid(name string, pid int) error {
	p := strconv.Itoa(pid)
	file, err := os.Create(models.FromRootDir("pid/" + name + ".pid"))
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
	file, err := os.Open(models.FromRootDir("pid/" + name + ".pid"))
	if err != nil {
		return err
	}
	reader := bufio.NewReader(file)
	pid, _, err := reader.ReadLine()
	if err != nil {
		return err
	}

	p, _ := strconv.Atoi(string(pid))
	process, err := os.FindProcess(p)
	if err != nil {
		os.Remove(models.FromRootDir("pid/" + name + ".pid"))
		return errFailedFindProcess
	}
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		os.Remove(models.FromRootDir("pid/" + name + ".pid"))
		return errProcessNotResponse
	}

	return nil
}
