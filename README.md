golang-db-pool-pattern
======================

An example of a DB connection pool written in Go using a master/worker pattern designed for processing large batches of operations.

It features:

 * Configurable number of workers (defaults to 1 per CPU core)
 * Configurable number of jobs
 * Progress output (in 5% chunks)
 * Summary statistics after all jobs are processed
 * Retry mechanism if DB connectivity is lost


Example: 
```bash
# Spawn a mongodb server for testing
$ mongod &

# Fetch this example
$ go get github.com/PaulMaddox/golang-db-pool-pattern

# See it in action
$ golang-db-pool-pattern --threads=16 --jobs=64000
2014/04/29 16:16:30 Running 64000 jobs across 8 workers
2014/04/29 16:16:30 Worker 0: Connecting to mongodb://localhost/worker-test
2014/04/29 16:16:30 Worker 1: Connecting to mongodb://localhost/worker-test
2014/04/29 16:16:30 Worker 2: Connecting to mongodb://localhost/worker-test
2014/04/29 16:16:30 Worker 3: Connecting to mongodb://localhost/worker-test
2014/04/29 16:16:30 Worker 4: Connecting to mongodb://localhost/worker-test
2014/04/29 16:16:30 Worker 5: Connecting to mongodb://localhost/worker-test
2014/04/29 16:16:30 Worker 6: Connecting to mongodb://localhost/worker-test
2014/04/29 16:16:30 Worker 7: Connecting to mongodb://localhost/worker-test
2014/04/29 16:16:30 Processing 5% complete
2014/04/29 16:16:30 Processing 10% complete
2014/04/29 16:16:30 Processing 15% complete
2014/04/29 16:16:31 Processing 20% complete
2014/04/29 16:16:31 Processing 25% complete
2014/04/29 16:16:31 Processing 30% complete
2014/04/29 16:16:31 Processing 35% complete
2014/04/29 16:16:31 Processing 40% complete
2014/04/29 16:16:31 Processing 45% complete
2014/04/29 16:16:32 Processing 50% complete
2014/04/29 16:16:32 Processing 55% complete
2014/04/29 16:16:32 Processing 60% complete
2014/04/29 16:16:32 Processing 65% complete
2014/04/29 16:16:32 Processing 70% complete
2014/04/29 16:16:33 Processing 75% complete
2014/04/29 16:16:33 Processing 80% complete
2014/04/29 16:16:33 Processing 85% complete
2014/04/29 16:16:33 Processing 90% complete
2014/04/29 16:16:33 Processing 95% complete
2014/04/29 16:16:33 Processing 100% complete
2014/04/29 16:16:33 Closing job queue and terminating workers
2014/04/29 16:16:33 All threads completed successfully in 3.263975746s
2014/04/29 16:16:33 Average speed of 50.999us per job
```
