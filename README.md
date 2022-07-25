# Test task from Xena Exchange

## Task

Write **single-thread** de-duplicating handler for _job(id, timestamp)_.  
Jobs ordered by timestamp.  
Deduplication by job.id in statically configured time window.

## Test setup

Benchmark test suite should generate identical load for each tested handler.  
Configurable deduplication window size, jobs rate borders and duplicate probability.  
Handler results correctness checks.  
Handlers resource consumption check.  
Randomization seed may be passed as _RAND_SEED_ env variable.

## Run bench

`go test -bench=. -cpu=1 -benchtime=100000x -benchmem`