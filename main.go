package main

import (
	"log"

	"github.com/kamil5b/basic-life-sim/internal/core"
)

func main() {
	if err := core.Run(); err != nil {
		log.Fatal(err)
	}
}
