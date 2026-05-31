package main

import (
	"basic-life-sim/internal/core"
	"log"
)

func main() {
	if err := core.Run(); err != nil {
		log.Fatal(err)
	}
}
