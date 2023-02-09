package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	// "context"
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// Imitation enumeration
// enumeration of state of Args
var STATE_OF_ARGS = StateOfArgs{}

type StateOfArgs struct {
}

func (s *StateOfArgs) RequestTask() string {
	return "RequestTask"
}

func (s *StateOfArgs) Finished() string {
	return "Finished"
}

// enumeration of taskType of Args
var TASK_TYPE = TaskType{}

type TaskType struct {
}

func (t *TaskType) Map() string {
	return "Map"
}

func (t *TaskType) Reduce() string {
	return "Reduce"
}

func (t *TaskType) Wait() string {
	return "Wait"
}

func (t *TaskType) Done() string {
	return "Done"
}

// request
type Args struct {
	State string
	Tsk   Task
}

// enumeration of state of Reply
var STATE_OF_REPLY = StateOfReply{}

type StateOfReply struct {
}

func (s *StateOfReply) TaskInfo() string {
	return "TaskInfo"
}

func (s *StateOfReply) Received() string {
	return "Received"
}

// reply
type Reply struct {
	State string
	Tsk   Task
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
