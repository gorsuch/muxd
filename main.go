package main

import "fmt"
import "net/http"
import "net/url"
import "os"
import "github.com/fzzbt/radix/redis"

func searchEnv(list []string) string {
	for _, x := range list {
		s := os.Getenv(x)
		if len(s) > 0 {
			return s
		}
	}

	return ""
}

func redisUrl() url.URL {
	s := searchEnv([]string{"REDIS_URL", "REDISTOGO_URL", "OPENREDIS_URL", "MYREDIS_URL", "REDISGREEN_URL", "REDISCLOUD_URL"})
	if s == "" {
		s = "redis://localhost:6379"
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
	if u.User != nil {

		pass, set := u.User.Password()
		if set {
			conf.Password = pass
		}
	}
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
			// spawn a new client since we'll be modifing pubsub behavior
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
