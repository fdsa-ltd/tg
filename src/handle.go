package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Host struct {
	Root      string
	Port      string
	Templates []string
	Routes    []Route
}
type Route struct {
	Name    string
	Uri     string
	Asserts []string
	Filters []string
}
type Plugin struct {
	Name  string
	Path  string
	Entry string
}

var (
	temp    *template.Template
	plugins []Plugin
)

func assert(asserts []string, r *http.Request) bool {
	// # time in from end
	// # host in key...
	// # mothod in key...
	// # path in key...
	// # ip in key...
	// # query has key[=v]...
	// # cookie has key[=v]...
	// # header has key[=v]...
	for _, assert := range asserts {
		m := strings.Split(assert, " ")

		switch m[0] {
		case "time":
			if false {
				return false
			}
		case "host":
			if !IsExits(r.Host, m[1:]) {
				return false
			}
		case "method":
			if !IsExits(r.Method, m[1:]) {
				return false
			}
		case "path":
			if !IsExits(r.URL.Path, m[1:]) {
				return false
			}

		case "ip":
			if !IsExits(r.RemoteAddr, m[1:]) {
				return false
			}
		case "query":
			if !strings.Contains(r.URL.RawQuery, m[1]) {
				return false
			}
		case "cookie":
			if !strings.Contains(r.Header.Get("cookie"), m[1]) {
				return false
			}
		case "header":
			if r.Header.Get(m[1]) == "" {
				return false
			}
		}
	}
	return true
}
func IsExits(input string, keys []string) bool {
	for _, item := range keys {
		if strings.Index(input, item) == 0 {
			return true
		}
	}
	return false
}

func Filter(filters []string, r *http.Request) {
	// path insert append remove
	// header k v
	// cookie k v
	for _, filter := range filters {
		m := strings.Split(filter, " ")
		switch m[0] {
		case "path":
			switch m[1] {
			case "insert":
				r.URL.Path = strings.Join(append(m[2:], r.URL.Path), "/")
			case "append":
				r.URL.Path = strings.Join(append([]string{r.URL.Path}, m[2:]...), "/")
			case "remove":
				path := strings.Split(r.URL.Path, "/")
				for _, s := range m[2:] {

					i, err := strconv.Atoi(s)
					if err != nil {
						log.Println(err)
						continue
					}
					if i == 0 {
						path = path[1:]
					} else {
						if i == -1 {
							path = path[0 : len(path)-1]
						} else {
							path = append(path[:i], path[i+1:]...)
						}
					}

				}

				r.URL.Path = strings.Join(path, "/")
				log.Println(r.URL.Path)
			}

		case "header":
			if len(m) == 3 {
				r.Header.Set(m[1], m[2])
			} else {
				r.Header.Del(m[1])
			}
		case "cookie":
			c := &http.Cookie{
				Name:     m[1],
				Value:    m[2],
				Path:     "/",
				Expires:  time.Unix(0, 0),
				HttpOnly: true,
			}
			if len(m) == 3 {
				c.Expires = time.Now().Add(1)
			}
			r.AddCookie(c)
		}

	}

}

var lock = sync.RWMutex{}

func (my *Host) init() {
	lock.Lock()
	if len(my.Templates) > 0 {
		list := make([]string, len(my.Templates))
		for key, value := range my.Templates {
			list[key] = my.Root + "/" + value
		}
		temp, _ = template.ParseFiles(list...)
		plugins = getPlugins(my.Root + "/apps")
	}
	lock.Unlock()
}

func getPlugins(path string) []Plugin {
	result := make([]Plugin, 0)
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		var plugin = &Plugin{}

		if _, err := os.Lstat(path + "/" + f.Name() + "/" + "app.json"); !os.IsNotExist(err) {
			fileData, err := ioutil.ReadFile(path + "/" + f.Name() + "/" + "app.json")
			if nil == err {
				_ = json.Unmarshal([]byte(fileData), plugin)
			}
		}
		if plugin.Path == "" {
			plugin.Path = f.Name()
		}
		if plugin.Entry == "" {
			plugin.Entry = "/apps/" + f.Name() + "/dist/js/app.js"
		}
		if plugin.Name == "" {
			plugin.Name = f.Name()
		}
		log.Printf("发现应用：%s\n", *plugin)
		result = append(result, *plugin)
	}

	return result
}
func (host *Host) watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					host.init()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	for _, v := range host.Templates {
		err = watcher.Add(host.Root + "/" + v)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = watcher.Add(host.Root + "/apps")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
func (my *Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if my.Routes != nil && len(my.Routes) > 0 {
		for _, router := range my.Routes {
			url, err := url.Parse(router.Uri)
			if err != nil {
				log.Println(err)
				continue
			}

			if assert(router.Asserts, r) {
				Filter(router.Filters, r)
				log.Println(router)
				r.Host = url.Host
				proxy := httputil.NewSingleHostReverseProxy(url)
				proxy.ServeHTTP(w, r)
				return
			}
		}
	}
	if temp != nil {
		var path = r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if _, err := os.Lstat(my.Root + r.URL.Path); os.IsNotExist(err) {
			path = "/index.html"
		}
		t := temp.Lookup(path[1:])
		if nil != t {
			t.Execute(w, plugins)
			return
		}
	}
	http.FileServer(http.Dir(my.Root)).ServeHTTP(w, r)
}
