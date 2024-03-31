package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"time"
)

//
// This program calculates the value of Pi using a Monte Carlo method.
// There two nice properties of this calculation method:
// - It is easy to parallelize, being a good example of a CPU bound task
// - Execution can pick up where it last stopped: previous results from previous runs can be incorporated into this run
//

// Storing previous iteration results here. Format is: [in, total]
// Use the last log line of the previous run to capture the values.
// NOTE: DO NOT USE the values in the "in/total (sum)". Use instead the values from "in/total (this run)" log line
var previousIterations = [][]int64{
	// 2024-03-31: long serial run when the algorithm was running on a single core
	{321228337250, 409000000000},
	// 2024-03-31: parallel runs
	{4398286571, 5600000000},
	{13194656076, 16800000000},
	{16964658903, 21600000000},
	{258867190854, 329600000000},
	{131318464019, 167200000000},
	{69115044519, 88000000000},
}

func main() {
	in, total := sumPreviousIterations()
	calcParallel(in, total)
}

func calcParallel(in int64, total int64) {
	if in < 0 {
		in = 0
	}
	if total < 0 {
		total = 0
	}

	var runIn int64
	var runTotal int64

	var checkpointPi *big.Rat = pi(in, total)
	var checkpointTotal int64 = total
	var checkpointTime = time.Now()

	var syncEvery int64 = 100_000_000
	parallelism := runtime.NumCPU()
	inC := make(chan int64, 1024)

	for i := 0; i < parallelism; i++ {
		go worker(syncEvery, inC)
	}

	timeFormat := "2006-01-02 15:04:05"

	for {
		for i := 0; i < parallelism; i++ {
			part := <-inC
			in += part
			runIn += part
		}
		part := syncEvery * int64(parallelism)
		total += part
		runTotal += part

		pi := pi(in, total)

		elapsed := time.Since(checkpointTime)
		throughput := big.NewRat(total-checkpointTotal, elapsed.Nanoseconds())
		throughput.Mul(throughput, big.NewRat(1_000_000, 1))

		deltaPi := big.NewRat(0, 1)
		deltaPi.Sub(pi, checkpointPi)
		deltaPi.Abs(deltaPi)

		fmt.Printf(
			"time: %v, pi: %s, delta: %s, in/total (sum): %d/%d, in/total (this run): %d/%d, iterations per sec: %sK\n",
			time.Now().UTC().Format(timeFormat), pi.FloatString(30), deltaPi.FloatString(30), in, total, runIn, runTotal, throughput.FloatString(0),
		)

		checkpointTime = time.Now()
		checkpointPi = pi
		checkpointTotal = total
	}
}

func worker(syncEvery int64, inChannel chan<- int64) {
	// runtime.LockOSThread()
	// defer runtime.UnlockOSThread()
	for {
		var in int64 = 0
		for i := int64(0); i < syncEvery; i++ {
			x := rand.Float64()
			y := rand.Float64()

			xSquared := x * x
			ySquared := y * y

			if xSquared+ySquared <= 1 {
				in++
			}
		}
		inChannel <- in
	}
}

func pi(in int64, total int64) *big.Rat {
	if total == 0 {
		total = 1
	}
	pi := big.NewRat(in, total)
	pi.Mul(pi, big.NewRat(4, 1))
	return pi
}

func sumPreviousIterations() (int64, int64) {
	var inSum int64
	var totalSum int64
	for _, item := range previousIterations {
		inSum += item[0]
		totalSum += item[1]
	}
	return inSum, totalSum
}
