package cmd

import (
	"flag"
	"fmt"
	"gonder/api"
	"gonder/campaign"
	"gonder/models"
	"gonder/utm"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Run starting gonder from command line
func Run() {
	var (
		configFile string
	)

	flag.StringVar(&configFile, "c", "./dist_config.toml", "Path to config file")
	flag.StringVar(&models.LogDir, "l", "./log", "Path to log folder")
	flag.StringVar(&models.ServerPem, "p", "./cert/server.pem", "Path to certificate pem file")
	flag.StringVar(&models.ServerKey, "k", "./cert/server.key", "Path to certificate key file")
	initDb := flag.Bool("i", false, "Initial database")
	initDbY := flag.Bool("iy", false, "Initial database without confirm")
	showVer := flag.Bool("v", false, "Prints version")
	flag.Parse()

	if *showVer {
		fmt.Printf("Gonder version: v%s\r\n\r\n", models.Version)
		os.Exit(0)
	}

	if err := os.MkdirAll(models.LogDir, os.ModePerm); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	l, err := os.OpenFile(filepath.Join(models.LogDir, models.MainLog+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening log file: %v", err)
	}
	defer l.Close()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(io.MultiWriter(l, os.Stdout))

	err = models.ReadConfig(configFile)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	err = models.ConnectDb()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	if *initDb || *initDbY {
		err = models.ReadConfig(configFile)
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
		err = models.InitDb(*initDbY)
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
		fmt.Println("Ok")
		os.Exit(0)
	}

	err = models.CheckDb()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	apiLog, err := models.NewLogger(models.APILog)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	go api.Run(apiLog)

	campLog, err := models.NewLogger(models.CampaignLog)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	go campaign.Run(campLog)

	utmLog, err := models.NewLogger(models.UTMLog)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	utm.Run(utmLog)
}
