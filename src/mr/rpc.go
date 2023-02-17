package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
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
	TaskTypeNone TaskType = iota
	TaskTypeMap
	TaskTypeReduce 
	TaskTypeSleep
	TaskTypeExit
)

// finished task
type TaskArgs struct {
	DoneType TaskType
	Id       int
	Files    []string
}

// new task
type TaskReply struct {
	Type    TaskType
	Id      int
	Files   []string
	NReduce int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	// 强制类型转换 获取当前进程的实际用户 ID 
	s += strconv.Itoa(os.Getuid())
	return s
}
