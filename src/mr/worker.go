package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"sort"
	"time"

	"log"
	"net/rpc"
	"os"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

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

	var newTask TaskReply
	var finishedTask TaskArgs = TaskArgs{DoneType: TaskTypeNone}
	// os.Exit(0)

	// Your worker implementation here.
	// showAndContinue("Worker", newTask, finishedTask)
	for {
		// showAndContinue("Worker", newTask, finishedTask)
		newTask = GetTask(&finishedTask)
		switch newTask.Type {
		case TaskTypeMap:
			f := newTask.Files[0]
			// read
			file, err := os.Open(f)
			if err != nil {
				log.Fatalf("cannot open %v", f)
			}
			defer file.Close()
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cannot read %v", f)
			}
			// map function
			intermediate := mapf(f, string(content))
			// []KeyValue数组
			// 8 --> 10
			byReduceFiles := make(map[int][]KeyValue)
			for _, kv := range intermediate {
				idx := ihash(kv.Key) % newTask.NReduce
				byReduceFiles[idx] = append(byReduceFiles[idx], kv)
			}
			files := make([]string, newTask.NReduce)
			// int --> []KeyValue
			// rpc simulation
			for reduceId, kvs := range byReduceFiles {
				filename := fmt.Sprintf("mr-%d-%d", newTask.Id, reduceId)
				ofile, _ := os.Create(filename)
				defer ofile.Close()
				// 模拟远程传输json数据
				// To write key/value pairs in JSON format
				enc := json.NewEncoder(ofile)
				for _, kv := range kvs {
					// write kv
					err := enc.Encode(&kv)
					if err != nil {
						log.Fatal()
					}
				}
				files[reduceId] = filename
			}
			// files: [mr-0-0 mr-0-1 mr-0-2 mr-0-3 mr-0-4 mr-0-5 mr-0-6 mr-0-7 mr-0-8 mr-0-9]
			finishedTask = TaskArgs{DoneType: TaskTypeMap, Id: newTask.Id,
				Files: files}
		case TaskTypeReduce:
			intermediate := []KeyValue{}
			for _, filename := range newTask.Files {
				file, err := os.Open(filename)
				if err != nil {
					log.Fatalf("cannot open %v", filename)
				}
				defer file.Close()
				dec := json.NewDecoder(file)
				for {
					var kv KeyValue
					if err := dec.Decode(&kv); err != nil {
						break
					}
					intermediate = append(intermediate, kv)
				}
			}
			sort.Sort(ByKey(intermediate))
			oname := fmt.Sprintf("mr-out-%d", newTask.Id)
			ofile, _ := os.Create(oname)
			defer ofile.Close()
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
				fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)
				i = j
			}
			finishedTask = TaskArgs{DoneType: TaskTypeReduce, Id: newTask.Id,
				Files: []string{oname}}
		case TaskTypeSleep:
			// ???
			time.Sleep(500 * time.Millisecond)
			finishedTask = TaskArgs{DoneType: TaskTypeNone}
			// show("finishedTask", finishedTask)
		case TaskTypeExit:
			return
		default:
			panic(fmt.Sprintf("unknow type: %v", newTask.Type))
		}
	}

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()

}

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
		fmt.Printf("args.X == %v, ", args.X)
		fmt.Printf("reply.Y == %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//

func GetTask(finishedTask *TaskArgs) TaskReply {
	reply := TaskReply{}

	// send rpc request, wait for the reply
	// Coordinator.GetTask: we need GetTask function in Coordinator
	ok := call("Coordinator.GetTask", finishedTask, &reply)
	if !ok {
		fmt.Printf("call failed!\n")
		os.Exit(0)
	}
	return reply
}

func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	// 远程程序调用
	// 拨号
	// DialHTTP connects to an HTTP RPC server at the specified network address listening on the default HTTP RPC path.
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()
	// fmt.Println(args)
	// os.Exit(0)
	// 调用具体方法
	// rpcname = Coordinator.GetTask
	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
