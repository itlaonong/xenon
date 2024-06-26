/*
 * RadonDB
 *
 * Copyright 2021 The RadonDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package ctl

import (
	"context"
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof"

	"xenon/server"
	"xenon/xbase/xlog"
	"xenon/xbase/xrpc"

	"github.com/ant0ine/go-json-rest/rest"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
}

// Admin tuple.
type Admin struct {
	log    *xlog.Log
	server *http.Server
	xenon  *server.Server
}

// NewAdmin creates the new admin.
func NewAdmin(log *xlog.Log, xenon *server.Server) *Admin {
	return &Admin{
		log:   log,
		xenon: xenon,
	}
}

// Start starts http server.
func (admin *Admin) Start() {
	api := rest.NewApi()
	router, err := admin.NewRouter()
	if err != nil {
		panic(err)
	}

	authMiddleware := &rest.AuthBasicMiddleware{
		Realm: "xenon zone",
		Authenticator: func(userId string, password string) bool {
			if userId == admin.xenon.MySQLAdmin() && password == admin.xenon.MySQLPasswd() {
				return true
			}
			return false
		},
	}
	api.Use(authMiddleware)

	api.SetApp(router)
	handlers := api.MakeHandler()
	admin.server = &http.Server{Addr: admin.xenon.PeerAddress(), Handler: handlers}

	go func() {
		l := admin.log
		l.Info("http.server.start[%v]...", admin.xenon.PeerAddress())

		ln, err := xrpc.SetListener(admin.server.Addr)
		if err != nil {
			l.Panic("%v", err)
		}

		if err := admin.server.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
			l.Panic("%v", err)
		}
	}()
}

// Stop stops http server.
func (admin *Admin) Stop() {
	l := admin.log
	err := admin.server.Shutdown(context.Background())
	if err != nil {
		return
	}
	l.Info("http.server.gracefully.stop")
}
