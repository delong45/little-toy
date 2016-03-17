package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var productMap = map[string]string{
	"011010": "AddFans",
	"011020": "TopFans",
	"01110":  "FanstopExtend",
	"01120":  "Apploft",
	"01300":  "Wax",
	"0180":   "Bidfeed",
	"0170":   "Brand",
}

type DruidQuery struct {
	QueryType    string                 `json:"queryType"`
	DataSource   string                 `json:"dataSource"`
	Intervals    [1]string              `json:"intervals"`
	Granularity  string                 `json:"granularity"`
	Filter       map[string]string      `json:"filter"`
	Dimensions   [1]string              `json:"dimensions"`
	Aggregations [1](map[string]string) `json:"aggregations"`
	Context      map[string]int         `json:"context"`
}

func QueryDruid(url string, jsonStr []byte) []byte {
	fmt.Println("URL:>", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response Body:", string(body))

	return body
}

func WriteGraphite(service string, msg string) {
	conn, err := net.Dial("tcp", service)
	if err != nil {
		panic(err)
	}
	_, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Write error: %s", err.Error())
	}
	defer conn.Close()
}

func CreateIntervals() string {
	start := time.Now().Add(-time.Minute * 10)
	end := start.Add(+time.Minute * 5)
	startStr := start.Format("2006-01-02T15:04:05+08:00")
	endStr := end.Format("2006-01-02T15:04:05+08:00")
	intervals := startStr + "/" + endStr
	return intervals
}

func CreateRequestJson() []byte {
	var query DruidQuery

	query.QueryType = "groupBy"
	query.DataSource = "uve_stat_report"
	query.Intervals[0] = CreateIntervals()
	query.Granularity = "all"
	query.Dimensions[0] = "product_r"
	query.Filter = make(map[string]string)
	query.Filter["type"] = "selector"
	query.Filter["dimension"] = "service_name"
	query.Filter["value"] = "main_feed"
	query.Aggregations[0] = make(map[string]string)
	query.Aggregations[0]["name"] = "count"
	query.Aggregations[0]["type"] = "doubleSum"
	query.Aggregations[0]["fieldName"] = "count"
	query.Context = make(map[string]int)
	query.Context["timeout"] = 3000000

	jsonStr, err := json.Marshal(query)
	if err != nil {
		fmt.Println("json err:", err)
	}

	return jsonStr
}

func CreateImpressJson() []byte {
	var query DruidQuery

	query.QueryType = "groupBy"
	query.DataSource = "bo_adid"
	query.Intervals[0] = CreateIntervals()
	query.Granularity = "all"
	query.Dimensions[0] = "type"
	query.Filter = make(map[string]string)
	query.Filter["type"] = "selector"
	query.Filter["dimension"] = "service_name"
	query.Filter["value"] = "main_feed"
	query.Aggregations[0] = make(map[string]string)
	query.Aggregations[0]["name"] = "count"
	query.Aggregations[0]["type"] = "doubleSum"
	query.Aggregations[0]["fieldName"] = "count"
	query.Context = make(map[string]int)
	query.Context["timeout"] = 3000000

	jsonStr, err := json.Marshal(query)
	if err != nil {
		fmt.Println("json err:", err)
	}

	return jsonStr
}

func RequestParse() {

}

func ImpressParse() {

}

func RequestMonitor() {
	url := "http://10.39.7.42:9082/druid/v2"
	jsonStr := CreateRequestJson()
	fmt.Println("Product Request Json String:", string(jsonStr))

	body := QueryDruid(url, jsonStr)
	fmt.Println(body)
}

func ImpressMonitor() {
	url := "http://172.16.89.128:8082/druid/v2"
	jsonStr := CreateImpressJson()
	fmt.Println("Product Impression Json String:", string(jsonStr))

	body := QueryDruid(url, jsonStr)
	fmt.Println(body)
}

func main() {
	for {
		RequestMonitor()
		ImpressMonitor()
		time.Sleep(time.Second * 5)
	}
}
