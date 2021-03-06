package main

import (
	"github.com/xiangli-cmu/raft"
)

func init() {
	raft.RegisterCommand(new(SetCommand))
}

type SetCommand struct {
	Key   string
	Value string
}

func (c *SetCommand) CommandName() string {
	return "bench:set"
}

func (c *SetCommand) Apply(s raft.Server) (interface{}, error) {
	return nil, nil
}
