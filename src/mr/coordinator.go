package mr

import (
	// "fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type MapTask struct {
	id      int       // task id
	file    string    // map task input file
	startAt time.Time // task start time (for timeout)
	done    bool      // is task finished
}

type ReduceTask struct {
	id      int       // task id
	files   []string  // reduce task input files (M files)
	startAt time.Time // task start time (for timeout)
	done    bool      // is task finished
}

type Coordinator struct {
	// Your definitions here.
	mutex        sync.Mutex   // lock to protect shared data below
	mapTasks     []MapTask    // all map tasks
	mapRemain    int          // of remaining map tasks
	reduceTasks  []ReduceTask // all reduce tasks
	reduceRemain int          // of remaining reduce tasks
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//

func (c *Coordinator) GetTask(args *TaskArgs, reply *TaskReply) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// show("GetTask", args, args.DoneType)
	switch args.DoneType {
	case TaskTypeMap:
		if !c.mapTasks[args.Id].done {
			c.mapTasks[args.Id].done = true
			for reduceId, file := range args.Files {
				if len(file) > 0 {
					c.reduceTasks[reduceId].files = append(c.reduceTasks[reduceId].files, file)
				}
			}
			c.mapRemain--
		}
	case TaskTypeReduce:
		if !c.reduceTasks[args.Id].done {
			c.reduceTasks[args.Id].done = true
			c.reduceRemain--
		}
	}
	now := time.Now()
	timeoutAgo := now.Add(-10 * time.Second)
	// allocate
	// order from map to reduce
	if c.mapRemain > 0 {
		for idx := range c.mapTasks {
			t := &c.mapTasks[idx]
			// 未完成任务且有10s没有启动 --> 重新分配
			// 未开始的任务或被crash干掉的任务 ???
			// if !t.done && t.startAt.Before(timeoutAgo)
			if !t.done && t.startAt.Before(timeoutAgo) {
				reply.Type = TaskTypeMap
				reply.Id = t.id
				reply.Files = []string{t.file}
				reply.NReduce = len(c.reduceTasks)
				t.startAt = now
				return nil
			}
		}
		reply.Type = TaskTypeSleep
	} else if c.reduceRemain > 0 {
		for idx := range c.reduceTasks {
			t := &c.reduceTasks[idx]
			if !t.done && t.startAt.Before(timeoutAgo) {
				reply.Type = TaskTypeReduce
				reply.Id = t.id
				reply.Files = t.files
				t.startAt = now
				return nil
			}
		}
		reply.Type = TaskTypeSleep
	} else {
		reply.Type = TaskTypeExit
	}
	return nil
}

func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	//The network must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket".
	// Listen announces on the local network address.
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.mapRemain == 0 && c.reduceRemain == 0
	// Your code here.
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	// files := pg*.txt --> 8
	// nReduce := 10

	// Your code here.
	// init
	c := Coordinator{
		mapTasks:     make([]MapTask, len(files)),
		reduceTasks:  make([]ReduceTask, nReduce),
		mapRemain:    len(files),
		reduceRemain: nReduce,
	}
	for i, f := range files {
		c.mapTasks[i] = MapTask{id: i, file: f, done: false}
	}
	for i := 0; i < nReduce; i++ {
		c.reduceTasks[i] = ReduceTask{id: i, done: false}
	}
	c.server()
	return &c
}
