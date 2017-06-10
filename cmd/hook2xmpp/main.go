package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/genofire/golang-lib/log"
	"github.com/mattn/go-xmpp"

	"dev.sum7.eu/genofire/hook2xmpp/circleci"
	configuration "dev.sum7.eu/genofire/hook2xmpp/config"
	"dev.sum7.eu/genofire/hook2xmpp/git"
	ownXMPP "dev.sum7.eu/genofire/hook2xmpp/xmpp"
)

func main() {
	configFile := "config.conf"
	flag.StringVar(&configFile, "config", configFile, "path of configuration file")
	flag.Parse()

	// load config
	config := configuration.ReadConfigFile(configFile)
	options := xmpp.Options{
		Host:          config.XMPP.Host,
		User:          config.XMPP.Username,
		Password:      config.XMPP.Password,
		NoTLS:         config.XMPP.NoTLS,
		Debug:         config.XMPP.Debug,
		Session:       config.XMPP.Session,
		Status:        config.XMPP.Status,
		StatusMessage: config.XMPP.StatusMessage,
	}
	client, err := options.NewClient()
	if err != nil {
		log.Log.Panic(err)
	}

	log.Log.Infof("Started hock2xmpp with %s", client.JID())

	client.SendHtml(xmpp.Chat{Remote: config.XMPP.StartupNotify, Type: "chat", Text: "startup of hock2xmpp"})
	go ownXMPP.Start(client)

	circleciHandler := circleci.NewHandler(client, config.Hooks)
	http.Handle("/circleci", circleciHandler)

	gitHandler := git.NewHandler(client, config.Hooks)
	http.Handle("/git", gitHandler)

	srv := &http.Server{
		Addr: config.WebserverBind,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	// Wait for system signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs

	client.SendHtml(xmpp.Chat{Remote: config.XMPP.StartupNotify, Type: "chat", Text: "stopped of hock2xmpp"})

	srv.Close()

	log.Log.Info("received", sig)
}
