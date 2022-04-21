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

package main

import 
(
	// "bytes"
	// "context"
  // "container/list"
	"fmt"
	// "io/ioutil"
	// "net/http"
	"os"
	// "strings"
	"sync"
	// "time"

	"github.com/showalter/bdws/internal/data"
)


// -- Internal Structs --------------------------------------------------------

/**
 * A worker an its associated mutex.
 **/
type ProtectedWorker struct
{
  worker data.Worker
  mutex* sync.Mutex
}


// -- Global Variables --------------------------------------------------------

var workers []ProtectedWorker

var jobChannel  = make(chan data.Job, 100)
var taskChannel = make(chan data.Job, 100)


// -- Internal Routines -------------------------------------------------------


/** -- usage() ----------------------------------------------------------------
 *  Prints out usage information for the program.
 *
 *  @param args Command line arguments
 ** ------------------------------------------------------------------------ */
func usage(args []string) {

  fmt.Printf("Usage: %s <port>\n", args[0]);

}



/** -- main() -----------------------------------------------------------------
 *  The entry point of the program.
 ** ------------------------------------------------------------------------ */
func main() {

  /* Parse command line arguments */
	var args []string = os.Args

  if len(args) != 2 {
    usage(args);
    os.Exit(1)
  }

}



