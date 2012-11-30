package main

import "fmt"
import "net/http"
import "os"
import "time"
import "github.com/fzzbt/radix/redis"

func main() {
	conf := redis.DefaultConfig()
	c := redis.NewClient(conf)
	defer c.Close()

	h := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// TODO return error if these are not set
			c.Publish(r.FormValue("channel"), r.FormValue("data"))
		}

		if r.Method == "GET" {
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
			sub.Subscribe("mux")

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
