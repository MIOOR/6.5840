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

type TaskType int

const (
	TASK_TYPE_OF_NONE = iota
	TASK_TYPE_OF_MAP
	TASK_TYPE_OF_REDUCE
	TASK_TYPE_OF_WAIT
)

// request
type Args struct {
	// TskType string
	// index   int
	TskType TaskType
	Files   []string
	ID      int
}

// reply
type Reply struct {
	// TskType string
	TskType TaskType
	// index   int
	ID      int
	Files   []string
	NReduce int
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
