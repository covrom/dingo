package model

import "github.com/globalsign/mgo"

var DBName = "dingoblog"

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
	shema_struct{"comments", mgo.Index{
		Key:        []string{"Id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}},
	shema_struct{"comments", mgo.Index{
		Key: []string{"-CreatedAt"},
	}},
	shema_struct{"comments", mgo.Index{
		Key: []string{"Approved"},
	}},
	shema_struct{"comments", mgo.Index{
		Key: []string{"Parent"},
	}},
	shema_struct{"comments", mgo.Index{
		Key: []string{"PostId"},
	}},
	shema_struct{"messages", mgo.Index{
		Key: []string{"-CreatedAt"},
	}},
	shema_struct{"messages", mgo.Index{
		Key: []string{"IsRead"},
	}},
}
