/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"xenon/build"
	"xenon/config"
	"xenon/ctl"
	"xenon/raft"
	"xenon/server"
	"xenon/xbase/xlog"
)

var (
	flagConf string
	flagRole string
)

func init() {
	flag.StringVar(&flagConf, "c", "", "xenon config file")
	flag.StringVar(&flagConf, "config", "", "xenon config file")
	flag.StringVar(&flagRole, "r", "", "role type:[LEADER|FOLLOWER|IDLE]")
	flag.StringVar(&flagRole, "role", "", "role type:[LEADER|FOLLOWER|IDLE]")
}

func main() {
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	var state raft.State
	flag.Parse()

	info := build.GetInfo()
	fmt.Printf("xenon:[%+v]\n", info)
	if flagConf == "" {
		fmt.Printf("usage: %s [-c|--config <xenon_config_file>]\nxenon:[%+v]\n",
			os.Args[0], info)
		os.Exit(1)
	}

	// config
	conf, err := config.LoadConfig(flagConf)
	if err != nil {
		log.Panic("xenon load config error [%v]", err)
	}

	// set log level
	log.SetLevel(conf.Log.Level)

	// set the initialization state
	switch flagRole {
	case "LEADER":
		state = raft.LEADER
	case "FOLLOWER":
		state = raft.FOLLOWER
	case "IDLE":
		state = raft.IDLE
	default:
		state = raft.UNKNOW
	}

	// build
	log.Info("main: tag=[%s], git=[%s], go version=[%s], build date=[%s]",
		info.Tag, info.Git, info.GoVersion, info.Time)
	log.Warning("xenon.conf.raft:[%+v]", conf.Raft)
	log.Warning("xenon.conf.mysql:[%+v]", conf.Mysql)
	log.Warning("xenon.conf.mysqld:[%+v]", conf.Backup)

	// server
	s := server.NewServer(conf, log, state)
	s.Init()
	s.Start()
	log.Info("xenon.start.success...")

	if conf.Server.EnableAPIs {
		// Admin portal.
		admin := ctl.NewAdmin(log, s)
		admin.Start()
		defer admin.Stop()
	}

	s.Wait()

	log.Info("xenon.shutdown.complete...")
}
