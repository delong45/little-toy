package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
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

type DruidRequestResult struct {
	Event     EventRequest `json:"event"`
	Timestamp string       `json:"timestamp"`
	Version   string       `json:"version"`
}

type DruidImpressionResult struct {
	Event     EventImpression `json:"event"`
	Timestamp string          `json:"timestamp"`
	Version   string          `json:"version"`
}

type EventRequest struct {
	Count     float64 `json:"count"`
	Product_r string  `json:"product_r"`
}

type EventImpression struct {
	Count float64 `json:"count"`
	Type  string  `json:"type"`
}

const DruidTimeFormat = "2006-01-02T15:04:05.000Z"
const DruidQueryInterval = 5

func QueryDruid(url string, jsonStr []byte) []byte {
	fmt.Println("Druid URL:>", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Client.Do Error:", err)
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
		fmt.Println("Dial Error:", err)
	}
	_, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Write Error:", err)
	}
	defer conn.Close()
}

func CreateIntervals() string {
	start := time.Now().Add(-time.Minute * DruidQueryInterval * 2)
	end := start.Add(+time.Minute * DruidQueryInterval)
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
		fmt.Println("Marshal Error:", err)
	}
	return jsonStr
}

func CreateImpressionJson() []byte {
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
		fmt.Println("Marshal Error:", err)
	}
	return jsonStr
}

func RequestParse(body []byte) string {
	var msg string
	var requests []DruidRequestResult

	err := json.Unmarshal(body, &requests)
	if err != nil {
		fmt.Println("Unmarshal Error:", err)
	}

	var druidTime string
	var timeStamp string
	key := "uve_stats.product.requests"

	if len(requests) > 0 {
		druidTime = requests[0].Timestamp
		t, _ := time.Parse(DruidTimeFormat, druidTime)
		t = t.Add(+time.Minute * DruidQueryInterval)
		timeStamp = fmt.Sprintf("%d", t.Unix())
	}

	m := make(map[string]int)
	for _, req := range requests {
		count := int(req.Event.Count)
		product_r := req.Event.Product_r
		array := strings.Split(product_r, "|")
		for _, product := range array {
			if _, ok := m[product]; ok {
				m[product] += count
			} else {
				m[product] = count
			}
		}
	}
	for k, v := range m {
		line := key + "." + k + " " + strconv.Itoa(v) + " " + timeStamp + "\n"
		msg += line
	}

	return msg
}

func ImpressionParse(body []byte) string {
	var msg string
	var impressions []DruidImpressionResult

	err := json.Unmarshal(body, &impressions)
	if err != nil {
		fmt.Println("Unmarshal Error:", err)
	}

	var druidTime string
	var timeStamp string
	boCount := 0
	key := "uve_stats.product.impressions"

	if len(impressions) > 0 {
		druidTime = impressions[0].Timestamp
		t, _ := time.Parse(DruidTimeFormat, druidTime)
		t = t.Add(+time.Minute * DruidQueryInterval)
		timeStamp = fmt.Sprintf("%d", t.Unix())
	}

	for _, imp := range impressions {
		var product string
		count := int(imp.Event.Count)
		productType := imp.Event.Type

		if p, ok := productMap[productType]; ok {
			product = p
		} else if strings.HasPrefix(productType, "ad_") {
			boCount += count
			continue
		} else {
			product = "empty"
		}
		line := key + "." + product + " " + strconv.Itoa(count) + " " + timeStamp + "\n"
		msg += line
	}

	if boCount > 0 {
		msg += key + ".bo" + " " + strconv.Itoa(boCount) + " " + timeStamp + "\n"
	}

	return msg
}

func RequestMonitor() {
	url := "http://10.39.7.42:9082/druid/v2"
	jsonStr := CreateRequestJson()
	fmt.Println("Request of Product Json String:", string(jsonStr))

	body := QueryDruid(url, jsonStr)
	msg := RequestParse(body)

	service := "10.77.96.122:2003"
	WriteGraphite(service, msg)
}

func ImpressionMonitor() {
	url := "http://172.16.89.128:8082/druid/v2"
	jsonStr := CreateImpressionJson()
	fmt.Println("Impression of Product Json String:", string(jsonStr))

	body := QueryDruid(url, jsonStr)
	msg := ImpressionParse(body)

	service := "10.77.96.122:2003"
	WriteGraphite(service, msg)
}

func main() {
	for {
		RequestMonitor()
		ImpressionMonitor()
		time.Sleep(time.Minute)
	}
}
