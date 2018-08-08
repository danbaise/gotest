package gateway

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
)

type Conf struct {
	BackendIps   []string
	LocalAddress string
}

func (c *Conf) Get(key string) interface{} {
	r := reflect.ValueOf(c)
	return r.Elem().FieldByName(key).Interface()
}

func Config() *Conf {
	return configWithFile(ConfigFile)
}

func configWithFile(path string) *Conf {
	file, _ := os.Open(path)
	defer file.Close()

	decoder := json.NewDecoder(file)
	c := Conf{}
	err := decoder.Decode(&c)
	if err != nil {
		log.Print("Error:", err)
	}
	return &c
}
