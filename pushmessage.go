package main

import (
	"fmt"
	"os"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"strings"
)

var (
	apiUrl     = "https://api.day.app/"
	handlerUrl = map[string]string{"jinse": "https://api.jinse.com/v4/live/list?limit=10&reading=false", "wallstreetcn": "https://api-prod.wallstreetcn.com/apiv1/content/lives?channel=global-channel&client=pc&limit=10", "laohu8": "https://www.laohu8.com/api/v1/news/live/list?pageSize=20"}
)

type configuration struct {
	Secret   string
	File_url string
}

func main() {
	file, _ := os.Open("config.json")
	defer file.Close()

	decoder := json.NewDecoder(file)
	conf := configuration{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("Error:", err)
	}
	//fmt.Println(conf.File_url)

	resp, err := http.Get(conf.File_url)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	keywords := strings.Split(string(body), ",")
	//	fmt.Println(keywords)
	for _, v := range keywords {
		fmt.Println(v)
	}
	for k,v :=range handlerUrl{
		fmt.Println(k,v)
	}

}
