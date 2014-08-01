package tbeer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Env struct {
	ServerPort     int
	ServerSecure   bool
	ServerCertFile string
	ServerKeyFile  string
	FacebookAppid  string
	FacebookSecret string
}

var GlobalEnv *Env

func LoadEnv() {
	data, err := ioutil.ReadFile("env.json")
	if err != nil {
		fmt.Println(err)
	}
	var env Env
	err = json.Unmarshal(data, &env)
	if err != nil {
		fmt.Println("Json unmarshal error in env.json:")
		fmt.Println(err)
	}
	GlobalEnv = &env
}
