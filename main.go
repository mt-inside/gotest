package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	workers := make([]chan bool, 0)
	go func() {
		<-sigs
		for _, w := range workers {
			w <- true
		}
	}()
	var wg sync.WaitGroup
	//mux0 := http.NewServeMux()
	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello %s\n", r.Header.Get("user-agent"))
	})
	//mux0.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "hello %s\n", r.Header.Get("user-agent"))
	//})
	//mux1.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "hello %s\n", r.Header.Get("user-agent"))
	//})
	s0 := &http.Server{
		Addr: ":8000",
		//	Handler: mux0,
	}
	s1 := &http.Server{
		Addr: ":8001",
		//	Handler: mux1,
	}
	log.Print("Binding")
	workers = append(workers, timedServe(&wg, s0))
	workers = append(workers, timedServe(&wg, s1))
	log.Print("Serving...")

	wg.Wait()
}

func timedServe(wg *sync.WaitGroup, s *http.Server) chan bool {
	wg.Add(1)

	d := make(chan bool)
	t := time.NewTimer(5 * time.Second)

	log.Print("Listening on ", s.Addr)
	go s.ListenAndServe()

	go func() {
		select {
		case <-t.C:
			log.Print("Server timed out")
		case <-d:
			log.Print("Graceful shutdown")
		}

		log.Print("Closing server on ", s.Addr)
		s.Close()
		wg.Done()
	}()

	return d
}
