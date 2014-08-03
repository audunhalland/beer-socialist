package main

import (
	"github.com/audunhalland/beer-socialist"
	"runtime"
)

func main() {
	tbeer.LoadEnv()
	tbeer.InitDB()

	if tbeer.IsDBEmpty() {
		tbeer.PopulateRandom()
	}

	runtime.GOMAXPROCS(runtime.NumCPU() - 2)

	tbeer.StartHttp()
}
