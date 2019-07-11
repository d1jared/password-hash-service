/*
DESCRIPTION
This project contains a simple service for hashing passwords. Given a password, the service will hash 
the password using SHA512 and convert it to base64 encoding.

The project was created as a coding exercise for JumpCloud.

ASSUMPTIONS

* The hashed passwords are persisted in-memory: there is no long-term persistence. When the service is 
  stopped, all hashed passwords are gone.

* The object identifiers are not random. They increment monotonically by 1 with every POST request. 
  When the service is restarted, the identifiers start at 1 again. This is not particularly useful 
  for a real production service, but it's a good exercise in thread locking.

* Since the passwords are persisted in-memory, the service is not designed to scale behind a load 
  balancer. You can only run one instance of the service.

* The service is not secure. The current implementation doesn't use TLS: passwords should always 
  be transported using TLS. Also, the endpoints don't include any token validation for 
  authentication/authorization.

* The service doesn't place any limits on the request body size. In general, there is an assumption 
  that the service will not be accessible by evil clients. Adding request limits is not too 
  difficult, if needed.

* The service status data is very limited. The status should include: P95 and P99 times, memory 
  usage/stress, cpu usage/stress.

* Life is to be enjoyed.

AUTHOR
David Jared
July 2019
*/
package main

import ( 
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"strconv"
	"time"
	"encoding/base64"
	"crypto/sha512"
	"sync"
)

// Server stats
type Stats struct {
	// total number of requests
	requests int64
	
	// total time of all requests in microseconds
	totalTime int64
	
	// Mutux to sync access
	lock sync.Mutex
}

// Thread-safe counter
type SafeCounter struct {
	counter int64
	lock sync.Mutex
}

// Map of password hashes
var m map[int64]string

// Server stats
var s Stats

// ID counter for hashes
var idCounter SafeCounter

// Shutdown flag
var isShutdown bool = false

// Used to track the total number of POSTs requests and the total execution time
func statsTracker(start time.Time) {
	s.lock.Lock()
	defer s.lock.Unlock()
	
	s.requests++
	s.totalTime = s.totalTime + time.Since(start).Nanoseconds() / 1000;
}

// Used to fetch the stats in a thread safe manner
func fetchStats() (int64, int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.requests, s.totalTime
}

// Handle the hash GET request
func fetchHash(response http.ResponseWriter, request *http.Request) {
	log.Print("GET hash");
	// Check for shutdown
	if(isShutdown == true) {
		http.Error(response, "503 Service unavailable.", http.StatusServiceUnavailable)
		log.Print("GET hash: 503 Service unavailable");
		return
	}
	
	// Only allow GET requests
	if request.Method != http.MethodGet {
		http.Error(response, "405 Method not allowed.", http.StatusMethodNotAllowed)
		log.Print("GET hash: 405 Method not allowed");
		return
	}
	
	// Parse the id from the path
	p := strings.Split(request.URL.Path, "/")
	var id int64
	var err error
	id, err = strconv.ParseInt(p[len(p)-1], 10, 64)
	if(err != nil) {
		http.Error(response, "400 Bad request.", http.StatusBadRequest)
		log.Print("GET hash: 400 Bad request, id=", id);
		return
	}
	
	log.Print("GET hash: id=", id);
	
	// lookup the hashed password
	hash := m[id]
	
	// Did we find the hashed password?
	if(hash == "") {
		http.Error(response, "404 Not found.", http.StatusNotFound)
		log.Print("GET hash: 404 Not found");
		return
	}
	
	fmt.Fprintf(response, "%s", hash)
	log.Print("GET hash: Done");
}

// Handle the hash POST request
func createHash(response http.ResponseWriter, request *http.Request) {
	log.Print("POST hash");
	
	// Check for shutdown
	if(isShutdown == true) {
		http.Error(response, "503 Service unavailable.", http.StatusServiceUnavailable)
		log.Print("POST hash: 503 Service unavailable");
		return
	}
	
	// Only allow POST requests
	if request.Method != http.MethodPost {
		http.Error(response, "405 Method not allowed.", http.StatusMethodNotAllowed)
		log.Print("POST hash: 405 Method not allowed");
		return
	}
	
	// Keep track of request stats
	defer statsTracker(time.Now())
	
	// Get the password from the POST form
	password := request.PostFormValue("password");
	if password == "" {
		http.Error(response, "400 Bad request.", http.StatusBadRequest)
		log.Print("POST hash: 400 Bad request");
		return
	}
	
	// Convert string to bytes
	passwordBytes := []byte(password)
	
	// Calculate SHA512 Hash
	hasher := sha512.New()
	hasher.Write(passwordBytes)
	
	// Convert to base 64 encoding
	hash64 := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	
	// Get ID
	id := idCounter.FetchAndIncrement();
	
	// delay storing the hash in the map for 5 second
	timer := time.NewTimer(5 * time.Second)
	go func() {
		// Wait 5 seconds
		<-timer.C
		
		// Store in map
		m[id] = hash64;
	}()
	
	fmt.Fprintf(response, "%d", id)
	
	log.Print("POST hash: id=", id);
	log.Print("POST hash: Done");
}

// Handle the GET status request
func statusHandler(response http.ResponseWriter, request *http.Request) {
	log.Print("GET stats");
	
	// Check for shutdown
	if(isShutdown == true) {
		http.Error(response, "503 Service unavailable.", http.StatusServiceUnavailable)
		log.Print("GET stats: 503 Service unavailable");
		return
	}
	
	// Only allow GET requests
	if request.Method != http.MethodGet {
		http.Error(response, "405 Method not allowed.", http.StatusMethodNotAllowed)
		log.Print("GET stats: 405 Method not allowed");
		return
	}
	
	// fetch the current status
	requests, totalTime := fetchStats();
	
	if requests == 0 {
		fmt.Fprintf(response, "{\"total\": 0, \"average\": 0}")
	} else {
		fmt.Fprintf(response, "{\"total\": %d, \"average\": %d}", requests, totalTime/requests)
	}
    
	log.Print("GET stats: Done");
}

// Handle the GET shutdown request
func shutdownHandler(response http.ResponseWriter, request *http.Request) {
	log.Print("GET shutdown");
	
	// Check for shutdown
	if(isShutdown == true) {
		http.Error(response, "503 Service unavailable.", http.StatusServiceUnavailable)
		log.Print("GET shutdown: 503 Service unavailable");
		return
	}
    
	isShutdown = true;
    
	// delay shutdown for 6 second: let any pending POSTs complete
	timer := time.NewTimer(6 * time.Second)
	go func() {
		// Wait 6 seconds
		<-timer.C
		log.Print("GET shutdown: Stopping hash service"); 
		os.Exit(0)
	}()
	
	log.Print("GET shutdown: Done");
}

// Increment the ID counter and return it's current value, in a threadsafe manner
func (c *SafeCounter) FetchAndIncrement() int64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.counter++;
	return(c.counter);
}

func main() {
	log.Print("Starting hash service"); 
	
	// Create the password map
	m = make(map[int64]string)
	
	// Initialize endpoint handlers
	http.HandleFunc("/hash", createHash);
	http.HandleFunc("/hash/", fetchHash);
	http.HandleFunc("/stats", statusHandler);
	http.HandleFunc("/shutdown", shutdownHandler);
	
	// Start the server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
