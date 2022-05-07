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
 *   0a. Why are brackets before the type name? []int
 *    0b. Why is the type after the name? WHY ARE THEY OPTIONAL?
 *	1. Externally used structs fields MUST begin with a capital letter
 *
 **/

package main

import (
	"bytes"
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

	"time"

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

/**
 * Represents a portion of a job.
 **/
type Task struct {
	JobId         int
	FileName      string
	Extension     string
	Code          []byte
	Parameterized bool
	Parameter     int
	Args          []string
}

// -- Global Variables --------------------------------------------------------
var server *http.Server
var workers = make(chan ProtectedWorker, MAX_WORKERS)

var jobChannel = make(chan data.Job, MAX_WORKERS)
var jobDone = make(chan []string, 10) /* Signals completion of tasks */
var jobsCompleted = 0

var taskChannel = make(chan Task, MAX_WORKERS)
var taskResults = make(chan string, MAX_WORKERS)

// -- Internal Routines -------------------------------------------------------

/** -- dispatch() -------------------------------------------------------------
 *  Dispatches a task to a worker.
 * @param task  The task to dispatch
 * @param pWorker  The worker to dispatch the task to
 ** ------------------------------------------------------------------------ */
func dispatch(task Task, pWorker ProtectedWorker) {
	fmt.Println("[Supervisor] Dispatching task.")

	/* Package and send the task to the worker */
	pWorker.mutex.Lock()

	end := task.Parameter - 1
	if task.Parameterized {
		end = task.Parameter + 1
	}

	jobBytes := data.JobToJson(data.Job{
		Id:             task.JobId,
		Time:           time.Now(),
		Machines:       1,
		ParameterStart: task.Parameter,
		ParameterEnd:   end,
		FileName:       task.FileName,
		Extension:      task.Extension,
		Code:           task.Code,
		Args:           task.Args,
		Nruns:          1,
	})

	/* Launch an asyncronous post request and cancel if it stops responding */
	/*
		context, cancelChannel := context.WithCancel(Context.Background())

		req, _ = http.NewRequestWithContext(context, "POST",
			"http://" + pWorker.worker.Hostname + "/newjob",
			 bytes.NewReader(jobBytes))

		req.Header.Add("Content-Type", "application/json")
	*/

	resp, err := http.Post("http://"+pWorker.worker.Hostname+"/newjob",
		"text/plain", bytes.NewReader(jobBytes))

	if err == nil {

		/* Put the bytes from the request into a file */
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		result := buf.String()
		taskResults <- result

	} else {
		/* If the request failed, put the task back in the queue */
		taskChannel <- task

		/* Yield so another worker can maybe get the request */
		time.Sleep(time.Millisecond * 250)
	}

	pWorker.mutex.Unlock()
}

/** - taskManager() ----------------------------------------------------------
 *  Handles task queue.
 ** ------------------------------------------------------------------------ */
func taskManager() {
	for {

		task := <-taskChannel
		worker := <-workers

		/* Dispatch the task to the worker */
		dispatch(task, worker)
	}
}

/** -- supervisor() ----------------------------------------------------------
 *  Splits jobs in the job channel into tasks and places them in the task
 *  queue.
 ** ------------------------------------------------------------------------ */
func supervisor() {
	/* Infinitely handle jobs from the job channel */
	for true {
		job := <-jobChannel
		fmt.Println("[Supervisor] Received a job.")

		start := 0
		end := 0
		step := 1
		param := false

		/* If the job is parameterized, make that many tasks */
		if job.ParameterEnd < job.ParameterStart {
			start = job.ParameterStart
			end = job.ParameterEnd
			param = true
		} else { /* Otherwise, run one on every currently available worker */
			start = 0
			end = len(workers)
		}

		/* Insert the tasks into the task channel */
		for i := start; i <= end; i += step {
			task := Task{
				JobId:         job.Id,
				FileName:      job.FileName,
				Extension:     job.Extension,
				Code:          job.Code,
				Parameterized: param,
				Parameter:     i,
				Args:          job.Args,
			}
			taskChannel <- task
		}

		/* Wait until all tasks are completed */
		for true {
			if len(taskResults) == start-end {
				break
			}
			time.Sleep(time.Second)
			fmt.Println("[Supervisor] Waiting for tasks to complete.")
		}

	}
}

/** -- job() ------------------------------------------------------------------
 *  Handles a job request.
 *  @param w  Write the reply into this writer
 *  @param r  Information about the request
 ** ------------------------------------------------------------------------ */
func job(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[Supervisor] Received a job request.")

	/* Parse the HTTP Request */
	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	job := data.JsonToJob(buf)

	runs := job.Nruns
	job.Nruns = 1

	/* Send the job to the job channel N times */
	for i := 0; i < runs; i++ {
		jobChannel <- job
	}
}

/** -- register() -------------------------------------------------------------
 *  Registers a worker with the supervisor by placing it into the priority
 *	queue of available workers.
 *  @param w  Write the reply into this writer
 *  @param r  Information about the request
 ** ------------------------------------------------------------------------ */
func register(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[Supervisor] Received a register request.")

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

	/* Create the worker struct and append it to the queue */
	worker := data.Worker{Id: 0, Busy: false, Hostname: reg.Hostname}
	protectedWorker := ProtectedWorker{worker, &sync.Mutex{}}
	workers <- protectedWorker

	/* Send a response to the worker  */
	w.Write(data.WorkerToJson(worker))
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

	/* Spawn a thread to handle jobs */
	go supervisor()
	go taskManager()

	/* Spawn a thread to handle the server */
	go func() {
		defer done.Done()
		fmt.Println("[Supervisor] Starting server on port " + port)
		os.Stdout.Sync()
		server.ListenAndServe()
	}()

	done.Wait()
	fmt.Println("\n----- Server stopped -----")

}
