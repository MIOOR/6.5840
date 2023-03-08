# MIT6.824 Lab 1: MapReduce

[TOC]



## Introduction

Implement：

- worker process

  can calls application Map and Reduce functions and handles reading and writing files

- coordinator process（Master）

  hands out tasks to workers and copes with failed workers



## Getting started

### [setup Go](https://pdos.csail.mit.edu/6.824/labs/go.html)

You should use [Go](http://www.golang.org/) 1.15 or any later version.

We strongly recommend that you work on the labs on your own machine.

### Fetch the initial lab software

```bash
git clone git://g.csail.mit.edu/6.824-golabs-2022 6.58240
cd 6.58240
ls
```

### File Structure

`src/main/mrsequential.go`：a simple sequential mapreduce implementation. It runs the maps and reduces one at a time, in a single process.

`mrapps/wc.go`：provide a couple of MapReduce applications: word-count

`mrapps/indexer.go`：has a text indexer

run word count sequentially as follows:

```bash
$ cd ~/6.5840
$ cd src/main
$ go build -buildmode=plugin ../mrapps/wc.go
$ rm mr-out* # 首次运行该仓库可忽略这条指令
$ go run mrsequential.go wc.so pg*.txt
$ more mr-out-0
A 509
ABOUT 2
ACT 8
...
```

`mrsequential.go` leaves its output in the file `mr-out-0`. The input is from the text files named `pg-xxx.txt`.



## Your Job ([moderate/hard](https://pdos.csail.mit.edu/6.824/labs/guidance.html))

To implement a distributed MapReduce：the coordinator and the worker.

Put your implementation in `mr/coordinator.go`, `mr/worker.go`, and `mr/rpc.go`.

Here's how to run your code on the word-count MapReduce application. First, make sure the word-count plugin is freshly built:

```bash
go build -buildmode=plugin ../mrapps/wc.go
```

In the `main` directory, run the coordinator.

```bash
rm mr-out*
go run mrcoordinator.go pg-*.txt
```

The `pg-*.txt` arguments to `mrcoordinator.go` are the input files; each file corresponds to one "split", and is the input to one Map task. 

In one or more other windows, run some workers:

```bash
go run mrworker.go wc.so
```

output in `mr-out-*`. 

When you've completed the lab, the sorted union of the output files should match the sequential output, like this:

```bash
$ cat mr-out-* | sort | more
A 509
ABOUT 2
ACT 8
...
```

supply a test script in `main/test-mr.sh`.

If you run the test script now, it will hang because the coordinator never finishes:

```bash
$ cd ~/6.824/src/main
$ bash test-mr.sh
*** Starting wc test
```

You can change `ret := false` to true in the Done function in `mr/coordinator.go` so that the coordinator exits immediately. Then:

```bash
$ bash test-mr.sh
*** Starting wc test.
sort: No such file or directory
cmp: EOF on mr-wc-all
--- wc output is not the same as mr-correct-wc.txt
--- wc test: FAIL
$
```

The test script expects to see output in files named `mr-out-X`, one for each reduce task. 

When you've finished, the test script output should look like this:

```bash
$ bash test-mr.sh
*** Starting wc test.
--- wc test: PASS
*** Starting indexer test.
--- indexer test: PASS
*** Starting map parallelism test.
--- map parallelism test: PASS
*** Starting reduce parallelism test.
--- reduce parallelism test: PASS
*** Starting crash test.
--- crash test: PASS
*** PASSED ALL TESTS
$
```



## A few rules:

- The map phase should divide the intermediate keys into buckets for `nReduce` reduce tasks, where `nReduce` is the number of reduce tasks -- argument that `main/mrcoordinator.go` passes to `MakeCoordinator()`. 
- The worker implementation should put the output of the X'th reduce task in the file `mr-out-X`.
- A `mr-out-X` file should contain one line per Reduce function output. The line should be generated with the Go `"%v %v"` format, called with the key and value.
- You can modify `mr/worker.go`, `mr/coordinator.go`, and `mr/rpc.go` and temporarily modify other files for testing, but make sure your code works with the original versions; we'll test with the original versions.
- The worker should put intermediate Map output in files in the current directory.
- `main/mrcoordinator.go` expects `mr/coordinator.go` to implement a `Done()` method that returns true when the MapReduce job is completely finished; at that point, `mrcoordinator.go` will exit.
- When the job is completely finished, the worker processes should exit. A simple way to implement this is to use the return value from `call()`: if the worker fails to contact the coordinator, it can assume that the coordinator has exited because the job is done, and so the worker can terminate too.



## Hints

- The [Guidance page](https://pdos.csail.mit.edu/6.824/labs/guidance.html) has some tips on developing and debugging.

- One way to get started is to modify `mr/worker.go`'s `Worker()` to send an RPC to the coordinator asking for a task. Then modify the coordinator to respond with the file name of an as-yet-unstarted map task. Then modify the worker to read that file and call the application Map function, as in `mrsequential.go`.

- The application Map and Reduce functions are loaded at run-time using the Go plugin package, from files whose names end in `.so`.

- If you change anything in the `mr/` directory, you will probably have to re-build any MapReduce plugins you use, with something like `go build -race -buildmode=plugin ../mrapps/wc.go`

- This lab relies on the workers sharing a file system. That's straightforward when all workers run on the same machine, but would require a global filesystem like GFS if the workers ran on different machines.

- A reasonable naming convention for intermediate files is `mr-X-Y`, where X is the Map task number, and Y is the reduce task number.

- The worker's map task code will need a way to store intermediate key/value pairs in files in a way that can be correctly read back during reduce tasks. One possibility is to use Go's `encoding/json` package. To write key/value pairs in JSON format to an open file:

  ```go
    enc := json.NewEncoder(file)
    for _, kv := ... {
      err := enc.Encode(&kv)
  ```

  and to read such a file back:

  ```go
    dec := json.NewDecoder(file)
    for {
      var kv KeyValue
      if err := dec.Decode(&kv); err != nil {
        break
      }
      kva = append(kva, kv)
    }
  ```

- The map part of your worker can use the `ihash(key)` function (in `worker.go`) to pick the reduce task for a given key.

- You can steal some code from `mrsequential.go` for reading Map input files, for sorting intermedate key/value pairs between the Map and Reduce, and for storing Reduce output in files.

- The coordinator, as an RPC server, will be concurrent; don't forget to lock shared data.

- Use Go's race detector, with `go build -race` and `go run -race`. `test-mr.sh` by default runs the tests with the race detector.

- Workers will sometimes need to wait, e.g. reduces can't start until the last map has finished. One possibility is for workers to periodically ask the coordinator for work, sleeping with `time.Sleep()` between each request. Another possibility is for the relevant RPC handler in the coordinator to have a loop that waits, either with `time.Sleep()` or `sync.Cond`. Go runs the handler for each RPC in its own thread, so the fact that one handler is waiting won't prevent the coordinator from processing other RPCs.

- The coordinator can't reliably distinguish between crashed workers, workers that are alive but have stalled for some reason, and workers that are executing but too slowly to be useful. The best you can do is have the coordinator wait for some amount of time, and then give up and re-issue the task to a different worker. For this lab, have the coordinator wait for ten seconds; after that the coordinator should assume the worker has died (of course, it might not have).

- If you choose to implement Backup Tasks (Section 3.6), note that we test that your code doesn't schedule extraneous tasks when workers execute tasks without crashing. Backup tasks should only be scheduled after some relatively long period of time (e.g., 10s).

- To ensure that nobody observes partially written files in the presence of crashes, the MapReduce paper mentions the trick of using a temporary file and atomically renaming it once it is completely written. You can use `ioutil.TempFile` to create a temporary file and `os.Rename` to atomically rename it.

- `test-mr.sh` runs all its processes in the sub-directory `mr-tmp`, so if something goes wrong and you want to look at intermediate or output files, look there. You can temporarily modify `test-mr.sh` to `exit` after the failing test, so the script does not continue testing (and overwrite the output files).

- `test-mr-many.sh` provides a bare-bones script for running `test-mr.sh` with a timeout (which is how we'll test your code). It takes as an argument the number of times to run the tests. You should not run several `test-mr.sh` instances in parallel because the coordinator will reuse the same socket, causing conflicts.

- Go RPC sends only struct fields whose names start with capital letters. Sub-structures must also have capitalized field names.

- When calling the RPC `call()` function, the reply struct should contain all default values. RPC calls should look like this:

  ```go
   reply := SomeType{}
    call(..., &reply)
  ```

  without setting any fields of reply before the call. If you pass reply structures that have non-default fields, the RPC system may silently return incorrect values.



## Coding

### MakeCoordinator() and Done()

![Lab1流程图-第 1 页.drawio](./MIT6.824%20Lab%201%20MapReduce/Lab1%E6%B5%81%E7%A8%8B%E5%9B%BE-%E7%AC%AC%201%20%E9%A1%B5.drawio.svg)

```go
type Coordinator struct {
	// Your definitions here.
	
    // map tasks
    // reduce tasks
    // number of unfinished task
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
// nReduce = 10
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
    
    //init MapTasks
    
    // fill some parameters in ReduceTasks
    
	c.server()
	return &c
}
```



```go
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	// ret := false
	ret := true

	// Your code here.
    // return true if state of maps of map task and reduce task and worker all are done else false
    

	return ret
}
```

### Woker()

![Lab1流程图-第 3 页.drawio](./MIT6.824%20Lab%201%20MapReduce/Lab1%E6%B5%81%E7%A8%8B%E5%9B%BE-%E7%AC%AC%203%20%E9%A1%B5.drawio.svg)

```go
// main/mrworker.go calls this function.
func Worker(mapf func(stri ng, string) []KeyValue,
	reducef func(string, []string) string) {
	// Your worker implementation here.
	
    // loop until no task
    
        // get a task from RPC

        // if task is maptask
            // 1. open file

            // 2. run mapf

            // 3. store kvs into different intermediate files by NReduce(use ihash() mod NReduce)

        // else if task is reducetask
            // 1. read all intermediate files and merge into kva

            // 2. sort kva by key

            // 3. run reducef

            // 4. store kvs in a file

        // else break loop if have no task
    
    	// fill the args
    
}
```

### RPC

![Lab1流程图-第 2 页.drawio](./MIT6.824%20Lab%201%20MapReduce/Lab1%E6%B5%81%E7%A8%8B%E5%9B%BE-%E7%AC%AC%202%20%E9%A1%B5.drawio.svg)

```go
// RPC information type
type Args struct{
    TaskType int
    files []string
} 

type Reply struct{
    TaskType int
    files []string
    NReduce int
}


// RPC handler
func (c *Coordinator) RPCHandler(args *Args, reply *Reply) error {
	// process args
    
    	// Skip if TaskType is NONE
    
    	// process MapTask
    		
    		// fill the ReduceTask Intermediate Files with MapTask Outputfiles
    
    	// process ReduceTask
    
    // process reply
    
    // get MapTask
    
    // get ReduceTask if MapTasks are finished
    
    // reply.TaskType = NONE if all tasks are done
    
	return nil
}
```
