package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	commit  = "unset"
	release = "unset"
)

func main() {
	log.Println("Starting the service...")

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Port is not set.")
	}

	isReady := &atomic.Value{}
	isReady.Store(false)
	go func() {
		log.Printf("Readyz probe is negative by default...")
		time.Sleep(5 * time.Second)
		isReady.Store(true)
		log.Printf("Readyz probe is positive.")
	}()

	http.HandleFunc("/", hello)
	http.HandleFunc("/v", version)
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/readyz", readyz(isReady))

	srv := &http.Server{
		Addr: ":" + port,
	}

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()
	log.Printf("The service is ready to listen and servei on :%s.", port)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	log.Print("The service is shutting down...")
	srv.Shutdown(context.Background())
	log.Print("Done.")

}

func hello(w http.ResponseWriter, r *http.Request) {
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
<title>app</title>
<style>
pre {
  font-size: 16px;
}
</style>
</head>
<body>
<pre>
   _____
  /  _  \ ______ ______
 /  /_\  \\____ \\____ \
/    |    \  |_> >  |_> >
\____|__  /   __/|   __/
        \/|__|   |__| %s
Pod hostname: %s
Node: %s
Client IP: %s
Current date and time: %s
</pre>
</body>
</html>
`

	clientIP := r.Header.Get("X-Real-Ip")
	dt := time.Now()
	filledTemplate := fmt.Sprintf(htmlTemplate, release, os.Getenv("HOSTNAME"), os.Getenv("NODE_NAME"), clientIP, dt.String())
	fmt.Fprint(w, filledTemplate)
}

func version(w http.ResponseWriter, _ *http.Request) {
	info := struct {
		Commit  string `json:"commit"`
		Release string `json:"release"`
	}{
		commit, release,
	}

	body, err := json.Marshal(info)
	if err != nil {
		log.Printf("Could not encode info data: %v", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readyz(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
