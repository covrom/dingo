package model

// import "github.com/globalsign/mgo"

// var shema_indexes = map[string]mgo.Index{
// 	"posts": mgo.Index{
// 		Key:        []string{"isbn"},
// 		Unique:     true,
// 		DropDups:   true,
// 		Background: true,
// 		Sparse:     true,
// 	},
// }

const schema = `
CREATE TABLE IF NOT EXISTS
posts (
  id                 integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  title              varchar(150) NOT NULL,
  slug               varchar(150) NOT NULL,
  markdown           text,
  html               text,
  image              text,
  featured           BOOLEAN,
  page               BOOLEAN,
  allow_comment      BOOLEAN,
  published          BOOLEAN,
  comment_num        integer NOT NULL DEFAULT '0',
  language           varchar(6) NOT NULL DEFAULT 'en_US',
  meta_title         varchar(150),
  meta_description   varchar(200),
  created_at         datetime NOT NULL,
  created_by         integer NOT NULL,
  updated_at         datetime,
  updated_by         integer,
  published_at       datetime,
  published_by       integer
);

CREATE TABLE IF NOT EXISTS
tokens (
  id          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  value       varchar(40) NOT NULL,
  user_id     integer UNIQUE,
  created_at  datetime,
  expired_at  datetime
);

CREATE TABLE IF NOT EXISTS
users (
  id               integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  name             varchar(150) NOT NULL,
  slug             varchar(150) NOT NULL,
  password         varchar(60) NOT NULL,
  email            varchar(254) NOT NULL,
  image            text,
  cover            text,
  bio              varchar(200),
  website          text,
  location         text,
  accessibility    text,
  status           varchar(150) NOT NULL DEFAULT 'active',
  language         varchar(6) NOT NULL DEFAULT 'en_US',
  last_login       datetime,
  created_at       datetime NOT NULL,
  created_by       integer NOT NULL,
  updated_at       datetime,
  updated_by       integer
);

CREATE TABLE IF NOT EXISTS
categories (
  id                integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  name              varchar(150) NOT NULL,
  slug              varchar(150) NOT NULL,
  description       varchar(200),
  parent_id         integer,
  meta_title        varchar(150),
  meta_description  varchar(200),
  created_at        datetime NOT NULL,
  created_by        integer NOT NULL,
  updated_at        datetime,
  updated_by        integer
);

CREATE TABLE IF NOT EXISTS
tags (
  id                integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  name              varchar(150) NOT NULL,
  slug              varchar(150) NOT NULL,
  hidden            boolean NOT NULL DEFAULT 0,
  created_at        datetime NOT NULL,
  created_by        integer NOT NULL,
  updated_at        datetime,
  updated_by        integer
);

CREATE TABLE IF NOT EXISTS
comments (
  id            integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  post_id       varchar(150) NOT NULL,
  author        varchar(150) NOT NULL,
  author_email  varchar(150) NOT NULL,
  author_avatar varchar(50)  NOT NULL,
  author_url    varchar(200) NOT NULL,
  author_ip     varchar(100) NOT NULL,
  created_at    datetime     NOT NULL,
  content       text         NOT NULL,
  approved      tinyint      NOT NULL DEFAULT '0',
  agent         varchar(255) NOT NULL,
  type          varchar(20),
  parent        integer,
  user_id       integer
);

CREATE TABLE IF NOT EXISTS
posts_tags (
  id       integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  post_id  integer NOT NULL,
  tag_id   integer NOT NULL
);

CREATE TABLE IF NOT EXISTS
posts_categories (
  id           integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  post_id      integer NOT NULL,
  category_id  integer NOT NULL
);

CREATE TABLE IF NOT EXISTS
settings (
  id          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  key         varchar(150) NOT NULL,
  value       varchar(20) NOT NULL,
  type        varchar(150) NOT NULL DEFAULT 'core',
  created_at  datetime NOT NULL,
  created_by  integer NOT NULL,
  updated_at  datetime,
  updated_by  integer
);

CREATE TABLE IF NOT EXISTS
roles (
  id           integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  name         varchar(150) NOT NULL,
  description  varchar(200),
  created_at   datetime NOT NULL,
  created_by   integer NOT NULL,
  updated_at   datetime,
  updated_by   integer
);

CREATE TABLE IF NOT EXISTS
messages (
  id           integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  type         varchar(20) NOT NULL,
  data         text NOT NULL,
  is_read      boolean NOT NULL default 0,
  created_at   datetime NOT NULL
);
`
