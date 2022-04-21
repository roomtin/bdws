// data package for saving objects
package data

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Client struct {
	Id   int
	Time time.Time
}

/**
 * Saves a client's information into json
 */
func ClientToJson(client Client) []byte {

	// Save c as json byte array
	b, err := json.Marshal(client)

	// Exit on error, otherwise return b
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	return b
}

/**
 * Saves a client's information into json
 */
func ClientDataToJson(id int, time time.Time) []byte {

	// Create Client Object
	c := Client{id, time}

	return ClientToJson(c)
}

/**
 * converts a []byte of json into a client struct
 */
func JsonToClient(b []byte) Client {
	var c Client

	// Unmarshall b into Client c
	err := json.Unmarshal(b, &c)

	// Exit on error, otherwise return c
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	return c
}

type Job struct {
	Id             int
	Time           time.Time
	Machines       int
	ParameterStart int
	ParameterEnd   int
	FileName       string
	Extension      string
	Code           []byte
	Args           []string
	Nruns          int
}

/**
 * Saves a Job information into json
 */
func JobToJson(job Job) []byte {

	// Save c as json byte array
	b, err := json.Marshal(job)

	// Exit on error, otherwise return b
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	return b
}

/**
 * Saves a Job information into json
 */
func JobDataToJson(id int, time time.Time, machines int,
	parameterStart int, parameterEnd int, fileName string, extension string, code []byte, args []string, nruns int) []byte {

	// Create Job Object
	j := Job{id, time, machines, parameterStart, parameterEnd, fileName, extension, code, args, nruns}

	return JobToJson(j)
}

/**
 * converts a []byte of json into a Job struct
 */
func JsonToJob(b []byte) Job {
	var j Job

	// Unmarshall b into Job j
	err := json.Unmarshal(b, &j)

	// Exit on error, otherwise return j
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	return j
}

type Worker struct {
	Id       int64
	Busy     bool
	Hostname string
}

/**
 * Saves a Worker information into json
 */
func WorkerToJson(worker Worker) []byte {

	// Save c as json byte array
	b, err := json.Marshal(worker)

	// Exit on error, otherwise return b
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	return b
}

/**
 * Saves a Worker information into json
 */
func WorkerDataToJson(id int64, busy bool, hostname string) []byte {

	// Create Worker Object
	w := Worker{id, busy, hostname}

	return WorkerToJson(w)
}

/**
 * Converts a []byte of json into a Worker struct
 */
func JsonToWorker(b []byte) Worker {
	var w Worker

	// Unmarshall b into Worker w
	err := json.Unmarshal(b, &w)

	// Exit on error, otherwise return j
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	return w
}

type Registration struct {
	Hostname     string
	Cores        int
	ModelName    string
	CpuSpeed     float64
	MemAvailable int
}

/** -- RegistrationDataToJson --------------------------------------------------
 * Saves a Registration information into json
 * @param registration Registration struct
 * @return []byte of json
 ** --------------------------------------------------------------------------*/
func RegistrationToJson(reg Registration) []byte {

	// Save c as json byte array
	b, err := json.Marshal(reg)

	// Exit on error, otherwise return b
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	return b
}

/** -- RegistrationToJson -------------------------------------------------------
 * Saves a Registration information into json
 * @param hostname string
 * @param cores int
 * @param model_name string
 * @param cpu_speed float64
 * @param mem_available int
 * @return []byte of json
 ** --------------------------------------------------------------------------*/
func RegistrationDataToJson(hostname string, cores int, model_name string, cpu_speed float64, mem_available int) []byte {

	// Create Registration Object
	r := Registration{hostname, cores, model_name, cpu_speed, mem_available}

	return RegistrationToJson(r)
}

/** -- JsonToRegistration -------------------------------------------------------
 * Converts a []byte of json into a Registration struct
 * @param b []byte
 * @return Registration struct
 ** --------------------------------------------------------------------------*/
func JsonToRegistration(b []byte) Registration {
	var r Registration

	// Unmarshall b into Registration r
	err := json.Unmarshal(b, &r)

	// Exit on error, otherwise return r
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	return r
}
