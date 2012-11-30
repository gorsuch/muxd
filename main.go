package main

import "fmt"
import "net/http"
import "net/url"
import "os"
import "github.com/fzzbt/radix/redis"

func redisUrl() url.URL {
	// TODO this is dumb.  be smarter.
	s := os.Getenv("REDIS_URL")
	if s == "" {
		s = os.Getenv("REDISTOGO_URL")
		if s == "" {
			s = os.Getenv("OPENREDIS_URL")
			if s == "" {
				s = os.Getenv("MYREDIS_URL")
				if s == "" {
					s = os.Getenv("REDISGREEN_URL")
					if s == "" {
						s = os.Getenv("REDISCLOUD_URL")
						if s == "" {
							s = "redis://localhost:6379"
						}
					}
				}
			}
		}
	}

	u, err := url.Parse(s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	return *u
}

// TODO support domain sockets
func redisConf() redis.Config {
	u := redisUrl()

	conf := redis.DefaultConfig()
	conf.Network = "tcp"
	conf.Address = u.Host
	return conf
}

func main() {
	c := redis.NewClient(redisConf())
	defer c.Close()

	h := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// TODO return error if these are not set
			// TODO do we need a mutex here?
			c.Publish(r.FormValue("channel"), r.FormValue("data"))
		}

		if r.Method == "GET" {
			c := redis.NewClient(redisConf())
			defer c.Close()

			// TODO allow channel size to be configurable
			lines := make(chan string)
			h := func(msg *redis.Message) {
				switch msg.Type {
				case redis.MessageMessage:
					lines <- msg.Payload
				}
			}

			sub, err := c.Subscription(h)
			if err != nil {
				panic(err)
			}
			defer sub.Close()
			sub.Subscribe(r.FormValue("channel"))

			for l := range lines {
				fmt.Fprintf(w, "%s\n", l)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
	}

	http.HandleFunc("/", h)
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
