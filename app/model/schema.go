package model

import "github.com/globalsign/mgo"

const DBName = "dingoblog"

type shema_struct struct {
	name string
	idx  mgo.Index
}

var shema_indexes = []shema_struct{
	shema_struct{"settings", mgo.Index{
		Key:        []string{"Key"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}},
}
