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
		fmt.Printf("Gonder version: %s (commit: %s date: %s)\r\n\r\n", models.AppVersion, models.AppCommit, models.AppDate)
		os.Exit(0)
	}

	if err := os.MkdirAll(models.LogDir, 0744); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := models.ReadConfig(configFile)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	l, err := models.NewLogfile("gonder_main")
	if err != nil {
		log.Printf("error opening log file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(io.MultiWriter(l, os.Stdout))

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

	models.InitPrometheus()

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
