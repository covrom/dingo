// Package Dingo is the main entry point to your application.
package Dingo

import (
	"fmt"

	"github.com/covrom/dingo/app/handler"
	"github.com/covrom/dingo/app/model"
	"github.com/dinever/golf"
)

// Init loads a public and private key pair used to create and validate JSON
// web tokens, or creates a new pair if they don't exist. It also initializes
// the database connection.
func Init(dbPath, privKey, pubKey string) {
	model.InitializeKey(privKey, pubKey)
	if err := model.Initialize(dbPath, false); err != nil {
		err = fmt.Errorf("failed to intialize db: %v", err)
		panic(err)
	}
	fmt.Printf("Database is used at %s\n", dbPath)
}

// Run starts our HTTP server on the given port.
func Run(portNumber string) {
	app := golf.New()
	app = handler.Initialize(app)
	fmt.Printf("Application Started on port %s\n", portNumber)
	app.Run(":" + portNumber)
}
