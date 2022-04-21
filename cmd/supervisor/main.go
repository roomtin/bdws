/**
 * This file contains the main routine for supervisors.
 *
 * This file was rewritten to have the notion of a job queue and a task queue.
 *
 * Jobs are split up into tasks (data decomposition) so that if workers disconnect
 * the task can be reinserted into the queue and completed by another worker, or
 * the same worker if it rejoins.
 *
 * Multiple jobs can now be queued up to be completed one after another.
 *
 * @author Jacob Bringham, Raleigh Martin, Parth Parikh
 * @version 4/16/2022
 **/

/**
 * Feel free to add anything to this list that bothers you about Go.
 *
 * Go Greivances:
 *  0. To write Go, just imagine that you are writing a program in C but backwards
	    0a. Why are brackets before the type name? []int
      0b. Why is the type after the name? WHY ARE THEY OPTIONAL?
 *	1. Externally used structs fields MUST begin with a capital letter
 *
 **/

package main

import (
	// "bytes"
	// "context"
	// "container/list"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"syscall"

	// "io/ioutil"
	// "net/http"
	"os"
	// "strings"

	"os/signal"
	"sync"

	// "time"

	"github.com/showalter/bdws/internal/data"
)

const MAX_WORKERS = 1000

// -- Internal Structs --------------------------------------------------------

/**
 * A worker an its associated mutex.
 **/
type ProtectedWorker struct {
	worker data.Worker
	mutex  *sync.Mutex
}

// -- Global Variables --------------------------------------------------------
var server *http.Server
var workers = make(PriorityQueue, MAX_WORKERS)
var jobChannel = make(chan data.Job, 100)
var taskChannel = make(chan data.Job, 100)

// -- Internal Routines -------------------------------------------------------

/** -- job() ------------------------------------------------------------------
 *  Handles a job request.
 *  @param w  Write the reply into this writer
 *  @param r  Information about the request
 ** ------------------------------------------------------------------------ */
func job(w http.ResponseWriter, r *http.Request) {

}

/** -- register() -------------------------------------------------------------
 *  Registers a worker with the supervisor by placing it into the priority
 *	queue of available workers.
 *  @param w  Write the reply into this writer
 *  @param r  Information about the request
 ** ------------------------------------------------------------------------ */
func register(w http.ResponseWriter, r *http.Request) {
	/* Parse the HTTP Request */
	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	reg := data.JsonToRegistration(buf)
	fmt.Printf("%+v\n", reg)

}

/** -- usage() ----------------------------------------------------------------
 *  Prints out usage information for the program.
 *
 *  @param args Command line arguments
 ** ------------------------------------------------------------------------ */
func usage(args []string) {
	fmt.Printf("Usage: %s <port>\n", args[0])
}

/** -- shutdown() -------------------------------------------------------------
 *  Shuts down the supervisor.
 ** ------------------------------------------------------------------------ */
func shutdown(sig os.Signal) {
	fmt.Println("\n----- Shutting down -----")
	server.Shutdown(context.Background())
}

/** -- main() -----------------------------------------------------------------
 *  The entry point of the program.
 ** ------------------------------------------------------------------------ */
func main() {

	/* Parse command line arguments */
	var args []string = os.Args

	if len(args) != 2 {
		usage(args)
		os.Exit(1)
	}

	port := args[0]

	/* Start the HTTP Server */
	server = &http.Server{Addr: port}
	http.HandleFunc("/job", job)
	http.HandleFunc("/register", register)

	done := &sync.WaitGroup{}
	done.Add(1)

	/* Install a signal handler to catch SIGINT and SIGTERM and shutdown gracefully */
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)
	go shutdown(<-signalChan)

	/* Spawn a thread to handle the server */
	go func() {
		defer done.Done()
		server.ListenAndServe()
	}()

	done.Wait()
	fmt.Println("\n----- Server stopped -----")

}
