package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

func main() {
	var host *Host
	conf := flag.String("c", "ng.yaml", "config file")
	flag.Parse()
	viper.SetConfigFile(*conf)
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if nil != err {
		log.Fatalln("ERROR:", err.Error())
		host = &Host{Root: "./public", Port: "5555", Templates: []string{"index.html"}}
	}

	err = viper.UnmarshalKey("host", &host)
	if nil != err {
		log.Fatalln("ERROR:", err.Error())
		host = &Host{Root: "./public", Port: "5555", Templates: []string{"index.html"}}
	}

	host.init()
	go host.watch()
	http.Handle("/", host)
	log.Printf("Listen On http://localhost:%s", host.Port)
	log.Printf("Root Directory: %s", host.Root)

	err = http.ListenAndServe(":"+host.Port, nil)
	if nil != err {
		log.Fatalln("ERROR:", err.Error())
	}
}
