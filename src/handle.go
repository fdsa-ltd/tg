package main

import (
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
	Name string
	Uri  string
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
			if strings.Index(r.URL.RawQuery, m[1]) < 0 {
				return false
			}
		case "cookie":
			if strings.Index(r.Header.Get("cookie"), m[1]) < 0 {
				return false
			}
		case "header":
			if false {
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
		// plugins = []Plugin{{Name: "app4", Uri: "/purehtml/js/app.js"}}
		plugins = getPlugins(my.Root + "/apps")
		log.Println(plugins)
	}
	lock.Unlock()
}
func getPlugins(path string) []Plugin {
	result := make([]Plugin, 0)
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		log.Printf("发现应用：%s\n", f.Name())
		result = append(result, Plugin{Name: f.Name(), Uri: "/apps/" + f.Name() + "/dist/js/app.js"})
	}
	// err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
	// 	if f == nil {
	// 		return err
	// 	}

	// 	if f.IsDir() {
	// 		log.Println(path)
	// 		log.Println(f.Name())
	// 		result = append(result, Plugin{Name: f.Name(), Uri: "/" + f.Name() + "/dist/js/app.js"})
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	fmt.Printf("filepath.Walk() returned %v\n", err)
	// }
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
	err = watcher.Add(host.Root)
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
		if r.URL.Path == "" || r.URL.Path == "/" {
			temp.Execute(w, plugins)
			return
		}
		// log.Printf("Root: %s Path: %s\n", "判断文件是否存在", r.URL.Path)
		if _, err := os.Stat(my.Root + r.URL.Path); os.IsNotExist(err) {
			// data, err := ioutil.ReadFile(this.Root + "/index.html")

			// if err == nil {
			// 	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
			// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
			// 	w.Write(data)
			// 	return
			// }

			temp.Execute(w, plugins)
			return
		}

	}
	http.FileServer(http.Dir(my.Root)).ServeHTTP(w, r)
}
