package model

import "github.com/globalsign/mgo"

// TODO: Autoincrement id | _id fields

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
		Key: []string{"Parent"},
	}},
	shema_struct{"comments", mgo.Index{
		Key: []string{"PostId", "Parent", "Approved"},
	}},
	shema_struct{"messages", mgo.Index{
		Key:        []string{"Id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}},
	shema_struct{"messages", mgo.Index{
		Key: []string{"IsRead"},
	}},
	shema_struct{"posts", mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}},
	shema_struct{"posts", mgo.Index{
		Key: []string{"slug"},
	}},
	shema_struct{"posts", mgo.Index{
		Key: []string{"id", "is_page", "published"},
	}},

	shema_struct{"posts_tags", mgo.Index{
		Key: []string{"post_id"},
	}},
	shema_struct{"posts_tags", mgo.Index{
		Key: []string{"tag_id"},
	}},
	
	shema_struct{"tags", mgo.Index{
		Key: []string{"slug"},
	}},

}
