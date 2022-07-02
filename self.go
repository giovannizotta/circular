package main

import "log"

type Self struct {
	Id string
}

func NewSelf() *Self {
	result := &Self{}
	result.Id = getId()
	return result
}

func getId() string {
	info, err := lightning.GetInfo()
	if err != nil {
		log.Fatalln(err)
	}
	return info.Id
}
