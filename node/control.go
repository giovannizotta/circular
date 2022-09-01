package node

import (
	"github.com/elementsproject/glightning/jrpc2"
)

type Stop struct {
	Message string `json:"message"`
}

func (s *Stop) Name() string {
	return "circular-stop"
}

func (s *Stop) New() interface{} {
	return &Stop{}
}

func (s *Stop) Call() (jrpc2.Result, error) {
	GetNode().Stopped = true
	return &Stop{Message: "circular has been stopped. New commands will not fire htlcs until resumed"}, nil
}

type Resume struct {
	Message string `json:"message"`
}

func (s *Resume) Name() string {
	return "circular-resume"
}

func (s *Resume) New() interface{} {
	return &Resume{}
}

func (s *Resume) Call() (jrpc2.Result, error) {
	GetNode().Stopped = false
	return &Stop{Message: "circular has been resumed"}, nil
}
