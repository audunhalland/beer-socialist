package tbeer

import (
	"encoding/json"
	"io/ioutil"
)

type Env struct {
	FacebookAppid  string
	FacebookSecret string
}

var GlobalEnv *Env

func LoadEnv() {
	data, _ := ioutil.ReadFile("env.json")
	var env Env
	json.Unmarshal(data, &env)
	GlobalEnv = &env
}
