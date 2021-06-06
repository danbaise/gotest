package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const DOMIN = "xxx"
const KEY = "xxxx"
const SECRET = "xxxx"
const VPS_IP = "xxxxxx"

var IP string

func main() {
	t := time.Tick(10 * time.Minute)
	for {
		ip := getIP()
		if ip != IP {
			IP = ip
			modifyDDNS(IP)
		}
		<-t
	}
}

func getIP() string {
	req, err := http.Get("https://api.ipify.org/")
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

func modifyDDNS(ip string) {
	apiUrl := fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records", DOMIN)
	client := &http.Client{}

	body := strings.NewReader(fmt.Sprintf(`[{
	"data": "%s",
	"name": "ddns",
	"ttl": 600,
	"type": "A"
}, {
	"data": "%s",
	"name": "vps",
	"ttl": 600,
	"type": "A"
}, {
	"data": "ns75.domaincontrol.com",
	"name": "@",
	"ttl": 3600,
	"type": "NS"
}, {
	"data": "ns76.domaincontrol.com",
	"name": "@",
	"ttl": 3600,
	"type": "NS"
}]`, ip, VPS_IP))
	req, err := http.NewRequest(http.MethodPut, apiUrl, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("accept", `application/json`)
	req.Header.Add("Content-Type", `application/json`)
	req.Header.Add("Authorization", fmt.Sprintf("sso-key %s:%s", KEY, SECRET))
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(ip, resp.Status, string(respBody))
}
