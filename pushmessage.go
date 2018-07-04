package main

import (
	"os"
	"encoding/json"
	"io"
	"net/http"
	"io/ioutil"
	"strings"
	"sync"
	"crypto/md5"
	"encoding/hex"
	"time"
)

var (
	apiUrl     = "https://api.day.app/"
	handlerUrl = map[string]string{"jinse": "https://api.jinse.com/v4/live/list?limit=10&reading=false", "wallstreetcn": "https://api-prod.wallstreetcn.com/apiv1/content/lives?channel=global-channel&client=pc&limit=10", "laohu8": "https://www.laohu8.com/api/v1/news/live/list?pageSize=10"}
)

type configuration struct {
	Secret   string `json:"secret"`
	File_url string `json:"file_url"`
}

type handler struct {
	Keywords string   `json:"keywords"`
	Time     []onTime `json:"time"`
}

type onTime struct {
	Date    string `json:"date"`
	Content string `json:"content"`
}

type jinse struct {
	List []jinseList `json:"list"`
}

type jinseList struct {
	Lives []jinseLives `json:"lives"`
}

type jinseLives struct {
	Content string `json:"content"`
}

type laohu8 struct {
	Data laohu8items `json:"data"`
}

type laohu8items struct {
	Items []laohu8item `json:"items"`
}

type laohu8item struct {
	Content string `json:"content"`
}

type wallstreetcn struct {
	Data wallstreetcnitems `json:"data"`
}

type wallstreetcnitems struct {
	Items []wallstreetcnitem `json:"items"`
}

type wallstreetcnitem struct {
	Content_text string `json:"content_text"`
}

var kw []string
var pushMessage map[string]string
var beenSend map[string]struct{}

const base_format = "2006-01-02 15:04:05"

func init() {
	pushMessage = make(map[string]string)
	beenSend = make(map[string]struct{})
}

func main() {
	aChan := make(chan int, 1)
	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				run()
			}
		}
	}()
	//定时提醒
	onTimeTask()
	//阻塞主线程
	<-aChan
}

func onTimeTask() {
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				timeTask()
			}
		}
	}()
}

func timeTask() {
	conf := getConf()
	url := apiUrl + conf.Secret
	getRemoteConf := getRemoteConf(conf.File_url)
	for _, v := range getRemoteConf.Time {
		loc, _ := time.LoadLocation("Asia/Chongqing")
		t, _ := time.ParseInLocation(base_format, v.Date, loc)
		if t.Unix() == time.Now().Unix() {
			postMessage(url, v.Content)
		}
	}
}

func getConf() configuration {
	conf := configuration{}
	err := getConfig(&conf)
	checkErr(err)
	return conf
}

func getRemoteConf(files string) handler {
	resp, err := http.Get(files)
	defer resp.Body.Close()
	checkErr(err)
	body, err := ioutil.ReadAll(resp.Body)

	handler := handler{}
	err = json.Unmarshal(body, &handler)
	checkErr(err)
	return handler
}

func run() {
	conf := getConf()
	getRemoteConf := getRemoteConf(conf.File_url)
	kw := strings.Split(getRemoteConf.Keywords, ",")
	content := getContent()

	for _, v := range content {
		func(s string) {
			for _, v := range kw {
				if strings.Index(s, v) != -1 {
					w := md5.New()
					io.WriteString(w, s) //将s写入到w中
					var m string
					str := []rune(s)
					length := len(str)
					if length >= 200 {
						m = string(str[0: 200])
					} else {
						m = string(str[0: length])
					}
					pushMessage[hex.EncodeToString(w.Sum(nil))] = m
				}
			}
		}(v)
	}

	url := apiUrl + conf.Secret
	for k, v := range pushMessage {
		if checkMessage(k) == true {
			postMessage(url, v)
			beenSend[k] = struct{}{}
		}
	}
}

func checkMessage(v string) bool {
	length := len(beenSend)
	if length >= 100 {
		for k, _ := range beenSend {
			delete(beenSend, k)
		}
	}
	if _, ok := beenSend[v]; !ok {
		return true
	} else {
		return false
	}

}

func postMessage(url, str string) {
	resp, err := http.Get(url + "/" + str)
	checkErr(err)
	defer resp.Body.Close()
}

func getContent() []string {
	var output []string
	var wg sync.WaitGroup
	for k, v := range handlerUrl {
		wg.Add(1)
		go func(k, v string) {
			resp, err := http.Get(v)
			checkErr(err)
			body, err := ioutil.ReadAll(resp.Body)
			switch k {
			case "jinse":
				content := jinse{}
				err = json.Unmarshal(body, &content)
				for _, li := range content.List {
					for _, lv := range li.Lives {
						output = append(output, lv.Content)
					}
				}
			case "laohu8":
				content := laohu8{}
				err = json.Unmarshal(body, &content)
				for _, li := range content.Data.Items {
					output = append(output, li.Content)
				}

			case "wallstreetcn":
				content := wallstreetcn{}
				err = json.Unmarshal(body, &content)
				for _, li := range content.Data.Items {
					output = append(output, li.Content_text)
				}
			default:
			}
			resp.Body.Close()
			defer wg.Done()
		}(k, v)
	}
	wg.Wait()
	return output
}

func getConfig(conf *configuration) error {
	file, _ := os.Open("config.json")
	defer file.Close()
	err := jsonEncode(file, conf)
	return err
}

func jsonEncode(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)
	return err
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
