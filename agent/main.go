package main

import (
	"log"

	"github.com/Humalect/humalect-core/agent/tasks"
	"github.com/Humalect/humalect-core/agent/utils"
)

// TODO handle all if err != nill with a webhook at backend and as this is going open source so the webhook should be configurable
func main() {
	config := utils.ParseCLIArguments()
	err := tasks.Deploy(config)
	if err != nil {
		log.Fatal(err)
	}
}
