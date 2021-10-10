package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	conf := flag.String("c", "tg.json", "config file")
	flag.Parse()
	fileData, err := ioutil.ReadFile(*conf)
	if nil != err {
		log.Fatalln("ERROR:", err.Error())
		return
	}
	host := &Host{}
	err = json.Unmarshal([]byte(fileData), host)
	if nil != err {
		log.Fatalln("ERROR:", err.Error())
		return
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
