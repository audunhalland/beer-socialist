package main

import (
	"github.com/audunhalland/beer-socialist"
)

func main() {
	tbeer.LoadEnv()
	tbeer.InitDB()

	if tbeer.IsDBEmpty() {
		tbeer.PopulateRandom()
	}

	tbeer.StartHttp()
}
