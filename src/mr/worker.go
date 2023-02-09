package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	// "strconv"
	"sort"
	"time"
)

// import "sort"

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	// get and process task from coordinator
	task := MyCall(Args{State: STATE_OF_ARGS.RequestTask()})
	for task.TskType != TASK_TYPE.Done() {
		switch task.TskType {
		case TASK_TYPE.Map():
			// open file
			file, err := os.Open(task.InputFile)
			if err != nil {
				log.Fatalf("cannot open %v", task.InputFile)
			}
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cannot read %v", task.InputFile)
			}
			file.Close()

			// run mapf
			kva := mapf(task.InputFile, string(content))

			// store kvs in an intermediate file
			fileOfIntermediate, err := os.Create(task.OutputFile)
			if err != nil {
				log.Fatalf("cannot creat %v", fileOfIntermediate)
			}

			enc := json.NewEncoder(fileOfIntermediate)
			for _, kv := range kva {
				err = enc.Encode(&kv)
			}
			fileOfIntermediate.Close()

		case TASK_TYPE.Reduce():
			ofile, _ := os.Create(task.OutputFile)

			//
			// call Reduce on each distinct key in intermediate[],
			// and print the result to mr-out-0.
			//

			// open file
			file, err := os.Open(task.InputFile)
			if err != nil {
				log.Fatalf("cannot open %v", task.InputFile)
			}

			// get intermediate kvs from file
			var intermediate []KeyValue
			dec := json.NewDecoder(file)
			for {
				var kv KeyValue
				if err := dec.Decode(&kv); err != nil {
					break
				}
				intermediate = append(intermediate, kv)
			}
			file.Close()
			// sort intermediate kvs by key
			sort.Sort(ByKey(intermediate))
			// run reduce task
			i := 0
			for i < len(intermediate) {
				j := i + 1
				for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, intermediate[k].Value)
				}
				output := reducef(intermediate[i].Key, values)

				// this is the correct format for each line of Reduce output.
				fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

				i = j
			}

			ofile.Close()

		case TASK_TYPE.Wait():
			// wait for a second before get a task
			time.Sleep(time.Second)
		case TASK_TYPE.Done():
			// all tasks are Finished
		}
		time.Sleep(time.Second * 2)
		task = MyCall(Args{State: STATE_OF_ARGS.Finished(), Tsk: task})
	}

	// ----- endline of write code-----

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
}

// steal from src/main/mrsequential.go
// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// the end of steal code

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// My implementation of RPC call
// request: first must be "RequestTask" or "Finished"( call STATE_OF_ARGS.xxx() )
// only type string below if the first parameter is "Finished"
// second type is "Map" or "Reduce"( call TASK_TYPE.xxx() )
// third type is a file name which have done
// only return value if taskState is "Request" else return nil
func MyCall(args Args) (t Task) {

	// declare a reply structure.
	reply := Reply{}

	// fmt.Printf("BeforeCall\nargs: %v\n", args)

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	if call("Coordinator.MyRPCHandler", &args, &reply) {
		switch reply.State {
		case STATE_OF_REPLY.TaskInfo():
			t = reply.Tsk
		case STATE_OF_REPLY.Received():
			// means all task has done
		}
	} else {
		fmt.Println("Call failed!")
	}

	// fmt.Printf("AfterCall\nreply: %v\n\n", reply)

	return
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
