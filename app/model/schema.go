package model

import "github.com/globalsign/mgo"

var DBName = "dingoblog"

type shema_struct struct {
	name string
	idx  mgo.Index
}

var shema_indexes = []shema_struct{
	shema_struct{"settings", mgo.Index{
		Key:        []string{"key"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}},

	shema_struct{"comments", mgo.Index{
		Key: []string{"parent"},
	}},
	shema_struct{"comments", mgo.Index{
		Key: []string{"postid", "parent", "approved"},
	}},

	shema_struct{"messages", mgo.Index{
		Key: []string{"isread"},
	}},

	shema_struct{"posts", mgo.Index{
		Key: []string{"tags"},
	}},
	shema_struct{"posts", mgo.Index{
		Key: []string{"slug"},
	}},
	shema_struct{"posts", mgo.Index{
		Key: []string{"_id", "ispage", "ispublished"},
	}},

	shema_struct{"tokens", mgo.Index{
		Key: []string{"value"},
	}},

	shema_struct{"users", mgo.Index{
		Key: []string{"slug"},
	}},
	shema_struct{"users", mgo.Index{
		Key: []string{"name"},
	}},
	shema_struct{"users", mgo.Index{
		Key: []string{"email"},
	}},
}
