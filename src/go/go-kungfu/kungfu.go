package main

import (
	"log"
	"os"
	"sync"

	"github.com/luomai/kungfu/src/go/algo"
	rch "github.com/luomai/kungfu/src/go/rchannel"
	"github.com/luomai/kungfu/src/go/wire"
)

// #include <mpi.h>
import "C"

func code(err error) int {
	if err == nil {
		return 0
	}
	// TODO: https://www.open-mpi.org/doc/v3.1/man3/MPI.3.php#sect4
	return 1
}

var (
	cluster *rch.ClusterSpec
	router  *rch.Router
	server  *rch.Server
)

func exitErr(err error) {
	log.Printf("exit on error: %v", err)
	os.Exit(1)
}

//export Go_Kungfu_Init
func Go_Kungfu_Init() int {
	log.Print("Go_Kungfu_Init")
	var err error
	cluster, err = rch.NewClusterSpecFromEnv()
	if err != nil {
		exitErr(err)
	}
	router, err = rch.NewRouter(cluster)
	if err != nil {
		exitErr(err)
	}
	server, err = rch.NewServer(router)
	if err != nil {
		exitErr(err)
	}
	go server.ListenAndServe()
	return 0
}

//export Go_Kungfu_Finalize
func Go_Kungfu_Finalize() int {
	server.Close()
	// TODO: check error
	return 0
}

func bcast(buffer []byte, count int, dtype C.MPI_Datatype, root int, name string) int {
	n := count * wire.MPI_Datatype(dtype).Size()
	var wg sync.WaitGroup
	myRank := cluster.MyRank()
	if myRank == root {
		for i, addr := range cluster.Peers {
			if i != root {
				wg.Add(1)
				func(addr rch.NetAddr) {
					m := rch.NewMessage(buffer[:n])
					router.Send(addr.WithName(name), *m)
					wg.Done()
				}(addr)
			}
		}
	} else {
		var m rch.Message
		addr := cluster.Peers[root]
		router.Recv(addr.WithName(name), &m)
		if int(m.Length) != n {
			panic("unexpected recv length")
		}
		copy(buffer[:n], m.Data)
	}
	wg.Wait()
	// TODO: check error
	return 0
}

//export Go_Kungfu_Negotiate
func Go_Kungfu_Negotiate(sendBuf []byte, recvBuf []byte, count int, dtype C.MPI_Datatype, op C.MPI_Op, name string) int {
	log.Printf("Go_Kungfu_Negotiate: %s, %d, %d", name, count, dtype)
	root := 0
	n := count * wire.MPI_Datatype(dtype).Size()

	copy(recvBuf[:n], sendBuf[:n])
	myRank := cluster.MyRank()

	if myRank == root {
		var lock sync.Mutex
		var wg sync.WaitGroup
		for i, addr := range cluster.Peers {
			if i != root {
				wg.Add(1)
				func(addr rch.NetAddr) {
					var m rch.Message
					router.Recv(addr.WithName(name), &m)
					if int(m.Length) != n {
						panic("unexpected recv length")
					}
					buf := m.Data
					lock.Lock()
					algo.AddBy(recvBuf[:n], buf, count, wire.MPI_Datatype(dtype), wire.MPI_Op(op))
					lock.Unlock()
					wg.Done()
				}(addr)
			}
		}
		wg.Wait()
	} else {
		addr := cluster.Peers[root]
		m := rch.NewMessage(sendBuf[:n])
		router.Send(addr.WithName(name), *m)
	}

	return bcast(recvBuf, count, dtype, root, name)
}