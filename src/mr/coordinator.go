package mr

import (
	"fmt"
	// "internal/itoa"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	// "golang.org/x/text/cases"
)

type Coordinator struct {
	// Your definitions here.
	mapTasks                                  Tasks
	reduceTasks                               Tasks
	unfinishedMapTasks, unfinishedReduceTasks int
	// wokersState     map[string][]WokerState
}

// return a task and modifies it to be Running if it exists else return empty task
func (c *Coordinator) Get(TskType string) (task Task) {
	// give a task and changes it to Running if it exists
	switch TskType {
	case TASK_TYPE.Map():
		task = c.mapTasks.Find(TASK_STATE.Pending())
		c.mapTasks.ModifyState(TASK_STATE.Running(), task.Index)
	case TASK_TYPE.Reduce():
		task = c.reduceTasks.Find(TASK_STATE.Pending())
		c.reduceTasks.ModifyState(TASK_STATE.Running(), task.Index)
	}

	// give status to a empty task
	if task.TskType == "" {
		if c.unfinishedReduceTasks > 0 {
			// Reduce task is unstarted
			task.TskType = TASK_TYPE.Wait()
		} else {
			// all tasks are finished
			task.TskType = TASK_TYPE.Done()
		}
	}
	return
}

// // return true if the task status is modified successfully else return false
// func (c *Coordinator) modifyState(state string, t Task) (result bool) {
// 	switch t.TskType {
// 	case TASK_TYPE.Map():
// 		result = c.mapTasks.modifyState(state, t.Index)
// 	case TASK_TYPE.Reduce():
// 		result = c.reduceTasks.modifyState(state, t.Index)
// 	default:
// 		result = false
// 	}
// 	return
// }

// mark the task status as finished
func (c *Coordinator) Finished(t Task) {
	switch t.TskType {
	case TASK_TYPE.Map():
		c.mapTasks.ModifyState(TASK_STATE.Finished(), t.Index)
		c.unfinishedMapTasks--
	case TASK_TYPE.Reduce():
		c.reduceTasks.ModifyState(TASK_STATE.Finished(), t.Index)
		c.unfinishedReduceTasks--
	}
}

type Task struct {
	Index      int
	ID         string
	TskType    string
	State      string
	InputFile  string
	OutputFile string
}

// enumeration of state of task
var TASK_STATE = TaskState{}

type TaskState struct {
}

func (ts *TaskState) Pending() string {
	return "Pending"
}

func (ts *TaskState) Running() string {
	return "Running"
}

func (ts *TaskState) Finished() string {
	return "Finished"
}

type WokerState struct {
}

type Tasks struct {
	Tasks []Task
}

// append a new task to the Tasks array
func (ts *Tasks) Append(t Task) {
	ts.Tasks = append(ts.Tasks, t)
}

// change the state of the task
func (ts *Tasks) ModifyState(state string, Index int) (result bool) {
	if len(ts.Tasks) > Index {
		ts.Tasks[Index].State = state
		result = true
	} else {
		result = false
	}
	return
}

// return the first matching task else return empty result
func (ts *Tasks) Find(state string) (task Task) {
	for _, t := range ts.Tasks {
		if t.State == state {
			task = t
			break
		}
	}

	return
}

func (ts *Tasks) Empty() bool {
	if len(ts.Tasks) == 0 {
		return true
	} else {
		return false
	}
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) MyRPCHandler(args *Args, reply *Reply) error {

	// fmt.Println("called MyRPCHandler()")

	switch args.State {
	case STATE_OF_ARGS.Finished():
		switch args.Tsk.TskType {
		case TASK_TYPE.Map():
			c.Finished(args.Tsk)
		case TASK_TYPE.Reduce():
			c.Finished(args.Tsk)
		}
		fallthrough
	case STATE_OF_ARGS.RequestTask():
		reply.State = STATE_OF_REPLY.TaskInfo()
		if c.unfinishedMapTasks != 0 {
			reply.Tsk = c.Get(TASK_TYPE.Map())
		} else if c.unfinishedReduceTasks != 0 {
			reply.Tsk = c.Get(TASK_TYPE.Reduce())
		} else {
			reply.State = STATE_OF_REPLY.Received()
		}
	}

	return nil
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
	// ret := true

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
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	// c.mapTasks
	// c.reduceTasks = make(Tasks)
	// init Tasks
	reduceID := 0
	intermediate := ""
	for i, filename := range files {
		// init single map task
		reduceID = ihash(filename) % nReduce
		// reduceID = i
		intermediate = fmt.Sprintf("mr-%d-%d.txt", i, reduceID)
		c.mapTasks.Append(
			Task{Index: i,
				ID:         strconv.Itoa(i),
				TskType:    TASK_TYPE.Map(),
				State:      TASK_STATE.Pending(),
				InputFile:  filename,
				OutputFile: intermediate})

		// init single reduce task
		c.reduceTasks.Append(
			Task{Index: i,
				ID:         strconv.Itoa(reduceID),
				TskType:    TASK_TYPE.Reduce(),
				State:      TASK_STATE.Pending(),
				InputFile:  intermediate,
				OutputFile: fmt.Sprintf("mr-out-%d.txt", i)})

		// count the number of Tasks
		c.unfinishedMapTasks += 1
		c.unfinishedReduceTasks += 1
	}
	// fmt.Println(c)
	// ----- endline of write code-----

	c.server()
	return &c
}
