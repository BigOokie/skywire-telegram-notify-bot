// Copyright © 2018 BigOokie
//
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package wcbotapp

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/BigOokie/skywire-wing-commander/internal/utils"
	"github.com/BigOokie/skywire-wing-commander/internal/wcconfig"
	"github.com/BigOokie/skywire-wing-commander/internal/wcconst"
	log "github.com/sirupsen/logrus"
)

var config wcconfig.Config
var dumpConfigFlag bool

// loadConfig manages the configuration load specifics
// offloading the detail from the `main()` funct
func loadConfig() (c wcconfig.Config) {
	log.Debugln("loadConfig: Start")
	defer log.Debugln("loadConfig: Complete")
	// Load configuration
	c, err := wcconfig.LoadConfigParameters("config", filepath.Join(utils.UserHome(), ".wingcommander"), map[string]interface{}{
		"telegram.debug":                 false,
		"monitor.intervalsec":            10,
		"monitor.heartbeatintmin":        120,
		"monitor.discoverymonitorintmin": 120,
		"skymanager.address":             "127.0.0.1:8000",
		"skymanager.discoveryaddress":    "discovery.skycoin.net:8001",
	})

	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
		return
	}
	return
}

func processCmdLineFlags() {
	var versionFlag, helpFlag, aboutFlag bool

	flag.BoolVar(&versionFlag, "v", false, "print current version")
	flag.BoolVar(&dumpConfigFlag, "config", false, "print current config")
	flag.BoolVar(&helpFlag, "help", false, "print application help")
	flag.BoolVar(&aboutFlag, "about", false, "print application information")
	flag.Parse()

	// if version cmd line flag `-v` then print version info and exit
	if versionFlag {
		fmt.Println(wcconst.BotAppVersion)
		fmt.Println("")
		os.Exit(0)
	}

	// if help cmd line flag `-help` then print version info and exit
	if helpFlag {
		fmt.Println(wcconst.MsgCmdLineHelp)
		fmt.Println("")
		os.Exit(0)
	}

	// if about cmd line flag `-about` then print version info and exit
	if aboutFlag {
		fmt.Println(wcconst.MsgAbout)
		fmt.Println("")
		os.Exit(0)
	}
}

func initLogging() {
	// Setup Log Formatter
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
}

func Run() {
	processCmdLineFlags()

	// Setup and initalise application logging
	initLogging()

	// Check and setup application instance control. Only allow a single instance to run
	appInstance := utils.InitAppInstance(wcconst.AppInstanceID)
	defer appInstance.TryUnlock()

	// Load configuration
	config := loadConfig()
	config.PrintConfig()
	if dumpConfigFlag {
		os.Exit(0)
	}

	// Setup OS Notification for Interupt or Kill signal - to cleanly terminate the app
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, os.Kill)

	log.Infoln("Skywire Wing Commander Telegram Bot - Starting.")
	defer log.Infoln("Skywire Wing Commander Telegram Bot - Stopped.")

	// Initiate a new Bot instance
	log.Infoln("Initiating Bot instance.")
	bot, err := NewBot(config)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infoln("Starting Bot instance.")
	go bot.Start()

	// Wait for the app to be signaled to terminate
	select {
	case signal := <-osSignal:
		if signal == os.Interrupt {
			log.Debugln(wcconst.MsgOSInteruptSig)
		} else if signal == os.Kill {
			log.Debugln(wcconst.MsgOSKillSig)
		}
	}
}