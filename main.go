// An example of a DB connection pool using a master/worker pattern
// to perform a batch of database operations with a retry mechanism,
// progress output and job statistics after all jobs have processed.
//
// This particular example uses MongoDB, however the pattern is
// not database specific.
//
// Author: Paul Maddox <paul.maddox@gmail.com>
// Date: April 2014

package main

import (
    "fmt"
    "io"
    "log"
    "math"
    "runtime"
    "time"

    "github.com/ogier/pflag"
    "labix.org/v2/mgo"
)

// User is our database collection structure
type User struct {
    Name    string `bson:"name"`
    Email   string `bson:"email"`
    Profile string `bson:"link"`
}

// Job structure holds details of each job
// This could be used to pass additional information to the worker
type Job struct {
    JobId int
}

// JobResult structure is returned by the worker to the master thread
// and contains information about whether the job was successful or not
type JobResult struct {
    JobId    int
    WorkerId int
    Error    error
}

// Allow our options to be configured as CLI parameters
var workers *int = pflag.Int("workers", runtime.NumCPU(), "The number of worker threads to spawn (default is 1 per CPU core)")
var jobs *int = pflag.Int("jobs", 128000, "The number of jobs to spawn")
var host *string = pflag.String("host", "localhost", "The MongoDB hostname to connect to")
var db *string = pflag.String("db", "worker-test", "The MongoDB database to use")

// Main spawns the required worker threads and then places all of the required
// work onto the work queue, where the workers will pick it up from
func main() {

    // Parse the CLI arguments
    pflag.Parse()

    log.Printf("Running %d jobs across %d workers", *jobs, *workers)

    // Setup buffered input/output queues for the workers
    queue := make(chan *Job, 512)
    results := make(chan *JobResult, 512)

    // Spin up the workers
    for id := 0; id < *workers; id++ {
        go worker(id, queue, results)
    }

    // Now that the workers are ready, start
    // a timer to see how long the processing takes
    start := time.Now()

    // Assign work to the workers
    // Do this in a new goroutine so that we don't block the results reading queue
    // if the queue hits it's buffer of 1024 items
    go func(jobs *int, queue chan<- *Job) {
        for i := 0; i < *jobs; i++ {
            queue <- &Job{JobId: i}
        }
    }(jobs, queue)

    // Get the results for each job
    announced := 0
    for i := 0; i < *jobs; i++ {

        // Announce progress percentage in 5% chunks
        percentage := int(math.Ceil(float64(i) / float64(*jobs) * 100))
        if percentage > announced {
            announced = percentage
            if percentage%5 == 0 {
                log.Printf("Processing %d%% complete", percentage)
            }
        }

        // Fetch a result from the results queue (blocking)
        result := <-results
        if result.Error != nil {
            log.Printf("Job %d failed on worker %d (%s)", result.JobId, result.WorkerId, result.Error)
            continue
        }

    }

    // We've got all of the results, so close the queue
    // which will terminate all of the workers
    log.Printf("Closing job queue and terminating workers")
    close(queue)

    duration := time.Now().Sub(start)
    ns := duration.Nanoseconds() / int64(*jobs)
    avg := time.Unix(0, ns).Sub(time.Unix(0, 0))

    log.Printf("All threads completed successfully in %s", duration.String())
    log.Printf("Average speed of %s per job", avg.String())

}

// Worker spawns a new worker process that connects to the DB
// and waits for incoming jobs in the 'queue' channel.
// If a job is successful it will send the results back on the 'results'
// channel. If a job fails to complete due to DB not being connected
// it will put the failed job back on the 'queue' channel, re-establish
// DB connectivity and the continue processing jobs.
func worker(id int, queue chan *Job, results chan<- *JobResult) {

    // Lets keep track of how many jobs this worker processed
    var count int64 = 0

    // Keep trying to connect to the database until we get a connection
    var session *mgo.Session
    users := connect(id, session)
    //defer session.Close()

    // Wait for incoming jobs on the job queue (blocking) or for the queue to close
    for job := range queue {

        // Perform the database query
        err := users.Insert(User{
            Name:    fmt.Sprintf("User %d", job.JobId),
            Email:   fmt.Sprintf("user-%d@example.com", job.JobId),
            Profile: fmt.Sprintf("http://example.com/%d", job.JobId),
        })

        if err == io.EOF || err == io.ErrUnexpectedEOF {
            // Our job hasn't completed because the database is no longer connected
            // Put our job back onto the queue (in another go routine to avoid blocking if queue buffer is full)
            // Then reconnect the database and continue processing
            go func(job *Job, queue chan *Job) {
                queue <- job
            }(job, queue)
            users = connect(id, session)
            continue
        }

        // Send our results back
        results <- &JobResult{
            JobId:    job.JobId,
            WorkerId: id,
            Error:    err,
        }

        count++

    }

}

// Connect (re)connects to the database and returns a handle to a mongodb
// collection which can be used for CRUD operations
func connect(workerId int, session *mgo.Session) *mgo.Collection {

    for {

        // Open a DB connection
        log.Printf("Worker %d: Connecting to %s", workerId, fmt.Sprintf("mongodb://%s/%s", *host, *db))
        s, err := mgo.Dial(*host)
        if err != nil {
            log.Printf("Worker %d: Unable to connect to database (%s)", workerId, err)
            continue
        }

        // Connect to the DB collection
        return s.DB(*db).C("users")

    }

}
