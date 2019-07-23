package cluster

import (
	"fmt"
	"github.com/hashicorp/raft"
	"io"
)

type RaftFSM struct {

}

func (fsm *RaftFSM)Apply(log *raft.Log) interface{}{
	fmt.Println("log: ", log)
	return nil
}

func (fsm *RaftFSM)Snapshot() (raft.FSMSnapshot, error){
	return nil, nil
}

func (fsm *RaftFSM)Restore(io.ReadCloser) error{
	return nil
}





type simpleFsm struct {

}

func (fsm *simpleFsm)Apply(log *raft.Log) interface{}{
	fmt.Println("log: ", log)
	return nil
}

func (fsm *simpleFsm)Snapshot() (raft.FSMSnapshot, error){
	return nil, nil
}

func (fsm *simpleFsm)Restore(io.ReadCloser) error{
	return nil
}