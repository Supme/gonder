package cmd

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"gonder/api"
	"gonder/campaign"
	"gonder/models"
	"gonder/utm"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

var (
	errFailedFindProcess  = errors.New("failed to find process")
	errProcessNotResponse = errors.New("process not response to signal")
)

// Run starting gonder from command line
func Run() {
	var (
		configFile string
		pidFile    string
	)

	flag.StringVar(&configFile, "c", "./config.ini", "Path to config file")
	start := flag.Bool("start", false, "Start as daemon")
	stop := flag.Bool("stop", false, "Stop daemon")
	restart := flag.Bool("restart", false, "Restart daemon")
	check := flag.Bool("status", false, "Check daemon status")
	flag.StringVar(&pidFile, "p", "pid/gonder.pid", "Path to pid file")
	initDb := flag.Bool("i", false, "Initial database")
	showVer := flag.Bool("v", false, "Prints version")
	flag.Parse()

	if *showVer {
		fmt.Printf("Gonder version: v%s\r\n\r\n", models.Version)
		os.Exit(0)
	}

	var err error

	if *initDb {
		err = models.Init(configFile)
		if err != nil{
			fmt.Print(err)
			os.Exit(1)
		}
		err = models.InitDb()
		if err != nil{
			fmt.Print(err)
			os.Exit(1)
		}
		fmt.Println("Ok")
		os.Exit(0)
	}

	if *check {
		err = checkPid(pidFile)
		if err == nil {
			fmt.Println("Gonder daemon is running")
		} else {
			fmt.Println("Gonder daemon is stoping")
		}
		os.Exit(0)
	}

	if *stop {
		err = stopProcess(pidFile)
		if err != nil {
			fmt.Println(err.Error())
		}
		os.Exit(0)
	}

	err = models.Init(configFile)
	if err != nil{
		fmt.Print(err)
		os.Exit(1)
	}


	if *start {
		err = startProcess(pidFile)
		if err != nil {
			fmt.Println(err.Error())
		}
		os.Exit(0)
	}

	if *restart {
		err = stopProcess(pidFile)
		if err != nil {
			fmt.Println(err.Error())
		}
		err = startProcess(pidFile)
		if err != nil {
			fmt.Println(err.Error())
		}
		os.Exit(0)
	}

	go api.Run()
	go campaign.Run()
	utm.Run()

}

func startProcess(pidFile string) error {
	err := checkPid(pidFile)
	if err == nil {
		return errors.New("Process " + pidFile + " already running")
	}
	p := exec.Command(os.Args[0])
	err = p.Start()
	if err != nil {
		return errors.New(pidFile + " start error: " + err.Error())
	}
	fmt.Println("Started pid", p.Process.Pid)
	err = setPid(pidFile, p.Process.Pid)
	if err != nil {
		return errors.New(pidFile + " set PID error: " + err.Error())
	}

	return nil
}

func stopProcess(pidFile string) error {
	err := checkPid(pidFile)
	if err != nil {
		return err
	}
	file, err := os.Open(pidFile)
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
		if err := os.Remove(pidFile); err != nil {
			log.Print(err)
		}
		return errFailedFindProcess
	}
	// ToDo wait
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	if err := os.Remove(pidFile); err != nil {
		log.Print(err)
	}
	fmt.Println("Process stoped")
	return nil
}

func setPid(pidFile string, pid int) error {
	p := strconv.Itoa(pid)
	file, err := os.Create(pidFile)
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

func checkPid(pidFile string) error {
	file, err := os.Open(pidFile)
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
		if err := os.Remove( pidFile); err != nil {
			log.Print(err)
		}
		return errFailedFindProcess
	}
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		if err := os.Remove(pidFile); err != nil {
			log.Print(err)
		}
		return errProcessNotResponse
	}

	return nil
}
