package mr

import (
	// "internal/itoa"
	// "fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	// "strconv"
	"sync"
	"time"
	// "golang.org/x/text/cases"
)

type Coordinator struct {
	// Your definitions here.
	mutex                                     sync.Mutex
	mapTasks                                  []MapTask
	reduceTasks                               []ReduceTask
	unfinishedMapTasks, unfinishedReduceTasks int
	maxTimeOfTask                             time.Duration
}

func (c *Coordinator) getMapTask() (mt MapTask, tskType TaskType) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Reset timeout tasks
	c.resetTasksOfTimeout(TASK_TYPE_OF_MAP)

	// Get the map task
	mt.State = TASK_STATE_OF_UNINIT
	for i := range c.mapTasks {
		task := &c.mapTasks[i]
		if task.State == TASK_STATE_OF_PENDING {
			task.State = TASK_STATE_OF_RUNNING
			task.StartTime = time.Now()
			mt = *task
			tskType = TASK_TYPE_OF_MAP
			break
		}
	}

	// return WAIT if still have tasks else return NONE
	if mt.State == TASK_STATE_OF_UNINIT {
		if c.unfinishedMapTasks > 0 {
			tskType = TASK_TYPE_OF_WAIT
		} else {
			tskType = TASK_TYPE_OF_NONE
		}
	}

	return
}

func (c *Coordinator) getReduceTask() (rt ReduceTask, tskType TaskType) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Reset timeout tasks
	c.resetTasksOfTimeout(TASK_TYPE_OF_REDUCE)

	// Get the map task
	rt.State = TASK_STATE_OF_UNINIT
	for i := range c.reduceTasks {
		task := &c.reduceTasks[i]
		if task.State == TASK_STATE_OF_PENDING {
			task.State = TASK_STATE_OF_RUNNING
			task.StartTime = time.Now()
			// fmt.Printf("task.files: %v\n", task.IntermediateFiles)
			rt = *task
			tskType = TASK_TYPE_OF_REDUCE
			break
		}
	}
	// fmt.Printf("A rt: %v\n", rt)
	// return wait or finished task status if task is empty
	if rt.State == TASK_STATE_OF_UNINIT {
		if c.unfinishedReduceTasks > 0 {
			tskType = TASK_TYPE_OF_WAIT
		} else {
			// return Done if ReduceTasks are all finished
			tskType = TASK_TYPE_OF_NONE
		}
	}
	// fmt.Printf("B rt: %v\n", rt)
	return
}

func (c *Coordinator) resetTasksOfTimeout(taskType int) {
	switch taskType {
	case TASK_TYPE_OF_MAP:
		for i := range c.mapTasks {
			task := &c.mapTasks[i]
			if task.State == TASK_STATE_OF_RUNNING && time.Since(task.StartTime) > c.maxTimeOfTask {
				task.State = TASK_STATE_OF_PENDING
			}
		}
	case TASK_TYPE_OF_REDUCE:
		for i := range c.reduceTasks {
			task := &c.reduceTasks[i]
			if task.State == TASK_STATE_OF_RUNNING && time.Since(task.StartTime) > c.maxTimeOfTask {
				task.State = TASK_STATE_OF_PENDING
			}
		}
	}
}

type TaskState int

const (
	TASK_STATE_OF_UNINIT = iota
	TASK_STATE_OF_PENDING
	TASK_STATE_OF_RUNNING
	TASK_STATE_OF_FINISHED
)

type MapTask struct {
	// Index     int
	ID int
	// State     string
	State     TaskState
	InputFile string
	// OutputFiles []string
	StartTime time.Time
	Cost      time.Duration
	NReduce   int
}

type ReduceTask struct {
	// Index      int
	ID int
	// State      string
	State             TaskState
	IntermediateFiles []string
	OutputFile        string
	StartTime         time.Time
	Cost              time.Duration
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) MyRPCHandler(args *Args, reply *Reply) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// process args
	switch args.TskType {
	case TASK_TYPE_OF_NONE:
		// worker just started, do Nothing
	case TASK_TYPE_OF_MAP:
		// jude task is finished
		if c.isFinished(TASK_TYPE_OF_MAP, args.ID) {
			// add intermediate Files to ReduceTasks
			// fmt.Printf("MapTask Output: %v\n", args.Files)
			for reduceID, file := range args.Files {
				task := &c.reduceTasks[reduceID]
				task.IntermediateFiles = append(task.IntermediateFiles, file)
			}
			// fmt.Printf("IntermediateFiles: %v\n", c.reduceTasks)
		}
	case TASK_TYPE_OF_REDUCE:
		// jude task is finished
		if c.isFinished(TASK_TYPE_OF_REDUCE, args.ID) {
			// nothing to do
		}
	}

	// process reply
	maptask, tskType := c.getMapTask()
	if tskType == TASK_TYPE_OF_NONE {
		// try to get ReduceTask if MapTasks are all finished
		var reducetask ReduceTask
		reducetask, tskType = c.getReduceTask()
		if tskType != TASK_TYPE_OF_NONE {
			// fmt.Printf("reduceTask: %v\n", reducetask)
			reply.TskType = TASK_TYPE_OF_REDUCE
			reply.ID = reducetask.ID
			reply.Files = reducetask.IntermediateFiles
			// fmt.Printf("task Type equal NONE?: %v\n", tskType == TASK_TYPE_OF_NONE)
		}
	} else if tskType != TASK_TYPE_OF_WAIT {
		// try to get MapTask
		reply.TskType = TASK_TYPE_OF_MAP
		reply.ID = maptask.ID
		// reply.Files = append(reply.Files, maptask.InputFile)
		reply.Files = []string{maptask.InputFile}
		reply.NReduce = maptask.NReduce
	}

	if tskType == TASK_TYPE_OF_NONE {
		// fmt.Printf("tskType is None? %v\n", tskType == TASK_TYPE_OF_NONE)
		// all tasks are finished
		reply.TskType = TASK_TYPE_OF_NONE
		// for _, task := range c.reduceTasks {
		// 	fmt.Printf("%v\n", task.IntermediateFiles)
		// }
		// fmt.Printf("reply: %v\n", reply)
	}

	return nil
}

// set the status of the task to finished if the task is finished else Pendging
func (c *Coordinator) isFinished(tskType TaskType, ID int) (result bool) {
	switch tskType {
	case TASK_TYPE_OF_MAP:
		task := &c.mapTasks[ID]
		if task.Cost = time.Since(task.StartTime); task.Cost <= c.maxTimeOfTask {
			task.State = TASK_STATE_OF_FINISHED
			c.unfinishedMapTasks--
			result = true
		} else {
			task.State = TASK_STATE_OF_PENDING
		}
	case TASK_TYPE_OF_REDUCE:
		task := &c.reduceTasks[ID]
		if task.Cost = time.Since(task.StartTime); task.Cost <= c.maxTimeOfTask {
			task.State = TASK_STATE_OF_FINISHED
			c.unfinishedReduceTasks--
			result = true
		} else {
			task.State = TASK_STATE_OF_PENDING
		}
	}

	return
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.
	// return true if all Tasks are done.
	if c.unfinishedMapTasks == 0 && c.unfinishedReduceTasks == 0 {
		ret = true
	}
	// ----- endline of write code-----

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce Tasks to use.
func MakeCoordinator(Files []string, nReduce int) *Coordinator {
	c := Coordinator{
		mapTasks:    make([]MapTask, len(Files)),
		reduceTasks: make([]ReduceTask, nReduce),
	}

	// Your code here.
	// init Tasks
	c.maxTimeOfTask = time.Second * 10

	// init single map task
	for i, filename := range Files {
		c.mapTasks[i] = MapTask{ID: i,
			State:     TASK_STATE_OF_PENDING,
			InputFile: filename,
			NReduce:   nReduce}

		// count the number of Tasks
		c.unfinishedMapTasks += 1
	}

	for i := 0; i < nReduce; i++ {
		// create a single reduce task but not init at all
		c.reduceTasks[i] = ReduceTask{ID: i, State: TASK_STATE_OF_PENDING}
		// count the number of Tasks
		c.unfinishedReduceTasks += 1
	}
	// ----- endline of write code-----

	c.server()
	return &c
}
