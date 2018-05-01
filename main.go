package main

import (
	"flag"

	"github.com/covrom/dingo/app"
)

func main() {
	portPtr := flag.String("port", "8000", "The port number to listen to.")
	dbUrlPtr := flag.String("database", "localhost", "The mongodb database url to use.")
	privKeyPathPtr := flag.String("priv-key", "blog.rsa", "The private key file path for JWT.")
	pubKeyPathPtr := flag.String("pub-key", "blog.rsa.pub", "The public key file path for JWT.")
	flag.Parse()

	Dingo.Init(*dbUrlPtr, *privKeyPathPtr, *pubKeyPathPtr)
	Dingo.Run(*portPtr)
}
