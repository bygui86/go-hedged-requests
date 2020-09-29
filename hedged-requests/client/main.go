package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	port = 8080

	urlInstanceA = "http://localhost:8090/ishealthy"
	urlInstanceB = "http://localhost:8091/ishealthy"
	urlInstanceC = "http://localhost:8092/ishealthy"

	// this value must be calculated against the server before running the client, for example using vegeta
	latency99percMillisec = 16
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/simple", simpleHandler).Methods(http.MethodGet)
	router.HandleFunc("/fanout", fanoutHandler).Methods(http.MethodGet)
	router.HandleFunc("/hedged", hedgedHandler).Methods(http.MethodGet)

	fmt.Printf("Open HTTP server on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	if err != nil {
		panic(err)
	}
}

func simpleHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Execute simple requests")
	executeSimple(urlInstanceA)
}

func fanoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Execute fanout requests")
	executeFanout([]string{urlInstanceA, urlInstanceB, urlInstanceC})
}

func hedgedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Execute hedged requests")
	executeHedged([]string{urlInstanceA, urlInstanceB, urlInstanceC})
}

/*
	This method simply issues an HTTP request to the server. This is the main function that does that.
	You can see that for each response received, we add a log entry.
	So it will be easy to track how many requests we issued afterwards.
*/
func executeSimple(url string) string {
	start := time.Now()
	response, _ := http.Get(url)
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("Request time: %d ms from url%s\n", time.Since(start).Nanoseconds()/time.Millisecond.Nanoseconds(), url)
	return fmt.Sprintf("%s from %s", body, url)
}

/*
	This method issues 3 HTTP requests to server for each request received (one request for each replica).
	As we only use the first one to come back we can see that the tail went down quite a bit
	(most of the time one of the replicas will respond in an acceptable time).
*/
func executeFanout(urls []string) string {
	ch := make(chan string, len(urls))
	for _, url := range urls {
		go func(u string) {
			ch <- executeSimple(u)
		}(url)
	}
	return <-ch
}

/*
	This method is very similar to the fanout implementation.
	The difference is that we are waiting 21 ms before issuing the subsequent requests.
	We don't cancel the any request as it may be a false positive, it may come back in 22 ms for example
	so we will use whatever response comes back first.
*/
func executeHedged(urls []string) string {
	ch := make(chan string, len(urls))
	for _, url := range urls {
		go func(u string, c chan string) {
			c <- executeSimple(u)
		}(url, ch)

		select {
		case r := <-ch:
			return r
		case <-time.After(latency99percMillisec * time.Millisecond):
		}
	}
	return <-ch
}
