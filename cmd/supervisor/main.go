// This file contains the main routine for supervisors.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/showalter/bdws/internal/data"
)

type ProtectedWorker struct {
	worker data.Worker
	mutex *sync.Mutex
}

var workers []ProtectedWorker
var workerCounter int64 = 1

var jobs []data.Job

func workerHandler(pWorker ProtectedWorker, job data.Job, args chan int64, results chan<- string) {
	
	for arg := range args {

		fmt.Printf("Starting an arg.\n")
		pWorker.mutex.Lock()

		job.ParameterStart = arg

		jobBytes := data.JobToJson(job)

		resp, err := http.Post("http://"+pWorker.worker.Hostname+"/newjob",
			"text/plain", bytes.NewReader(jobBytes))
		if err == nil {
			// Put the bytes from the request into a file
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)

			fmt.Printf("About to write to channel!\n")
			results <- buf.String()

		} else {
			fmt.Printf("Error!!!!!\n")
			// Write the argument back to the channel so this worker can try it again
			// or another worker can try it.
			args <- arg

			// Wait a bit of time so another worker has a chance to take the job if one
			// is available. We don't want this worker to fail a job and pick it up
			// immediately for it to fail again.
			time.Sleep(time.Millisecond * 250)
		}

		fmt.Printf("Done with an arg.\n")

		pWorker.mutex.Unlock()
	}
}

// Handle the submission of a new job.
func new_job(w http.ResponseWriter, req *http.Request) {

	fmt.Println("Handling connection...")

	// Parse the HTTP request.
	if err := req.ParseForm(); err != nil {
		panic(err)
	}

	// Put the bytes from the request into a file
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	job := data.JsonToJob(buf)

	args := make(chan int64, job.ParameterEnd - job.ParameterStart + 1)
	results := make(chan string)

	var responses string

	for _, w := range workers {
		go workerHandler(w, job, args, results)
	}

	for i := job.ParameterStart; i <= job.ParameterEnd; i++ {
		args <- i
	}

	for i := job.ParameterStart; i <= job.ParameterEnd; i++ {
		responses += <-results
		fmt.Printf("Got some results\n")
	}

	close(args)

	// Send a response back.
	w.Write([]byte(responses))
}

// Look through a list and add the item if it isn't in the list already.
// This is slow for big lists, but there won't likely be a large number of workers.
func appendIfUnique(list []ProtectedWorker, w ProtectedWorker) []ProtectedWorker {
	for _, x := range list {
		if x.worker.Hostname == w.worker.Hostname {
			return list
		}
	}

	return append(list, w)
}

// Handle a worker registering to receive work
func register(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling registration...")

	// Parse the HTTP request.
	if err := req.ParseForm(); err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)

	// The worker will send the port they'll be listening on
	port := buf.String()

	// Replace the port number for the sender
	split := strings.Split(req.RemoteAddr, ":")
	split[len(split)-1] = port

	worker := data.Worker{Id: workerCounter, Busy: false,
		Hostname: strings.Join(split, ":")}
	pWorker := ProtectedWorker{worker, &sync.Mutex{}}

	workerCounter += 1

	// We don't need multiple workers with the same hostname.
	workers = appendIfUnique(workers, pWorker)

	w.Write(data.WorkerToJson(worker))
}

// The entry point of the program.
func main() {

	// The command line arguments. args[0] is the port to run on.
	args := os.Args

	// If the right number of arguments weren't passed, ask for them.
	if len(args) != 2 {
		fmt.Println("Please pass the port to run on preceded by a colon. eg. :4001")
		os.Exit(1)
	}

	// If there is a request for /newjob,
	// the new_job routine will handle it.
	http.HandleFunc("/newjob", new_job)

	// Handle requests for /register with the register function
	http.HandleFunc("/register", register)

	// Listen on a port.
	http.ListenAndServe(args[1], nil)
}
