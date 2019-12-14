package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Message は送信するメッセージを保持する構造体
type Message struct {
	Text string `json:"text"`
}

// PostMessages は送信するメッセージ群を保持する構造体
type PostMessages struct {
	Messages []Message `json:"messages"`
}

var (
	// SlackURL は Slack チャンネルの URL
	SlackURL = os.Getenv("SLACK_URL")
)

func postMessage(msg Message) {
	jsonByte, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	req, err := http.NewRequest(
		"POST",
		SlackURL,
		bytes.NewBuffer(jsonByte),
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer resp.Body.Close()
	return
}

func makeMessages() PostMessages {
	var messages PostMessages
	doc, err := goquery.NewDocument("https://www.hankyu.co.jp")
	if err != nil {
		fmt.Printf("%s", err)
	}
	doc.Find("#current_status>table>tbody>tr>td[class!='current_status_link']>a").Each(func(i int, s *goquery.Selection) {
		body := s.Text()
		href, _ := s.Attr("href")
		messages.Messages = append(messages.Messages, Message{Text: body + "\n" + href})
	})
	return messages
}

func isJSON(fileName string) bool {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return false
	}
	return true
}

func readJSON(fileName string) PostMessages {
	var messages PostMessages
	raw, _ := ioutil.ReadFile(fileName)
	json.Unmarshal(raw, &messages)
	return messages
}

func isExistMsg(inspectMsg Message, msgs PostMessages) bool {
	for _, m := range msgs.Messages {
		if m.Text == inspectMsg.Text {
			return true
		}
	}
	return false
}

func writeJSON(fileName string, msgs PostMessages, msg Message) {
	msgs.Messages = append(msgs.Messages, msg)
	jsonByte, err := json.Marshal(msgs)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	ioutil.WriteFile(fileName, jsonByte, os.ModePerm)
}

func main() {
	// ServiceJSONPath は同じ内容の遅延情報を送信しないために使用する json ファイルのパス
	var ServiceJSONPath = flag.String("json-path", "./json/service.json", "遅延情報が記載された json を保存するファイルパス")

	flag.Parse()

	msgs := makeMessages()
	for _, msg := range msgs.Messages {
		if strings.Index(msg.Text, "発生しています") != -1 {
			if isJSON(*ServiceJSONPath) {
				messages := readJSON(*ServiceJSONPath)
				if !isExistMsg(msg, messages) {
					postMessage(msg)
					writeJSON(*ServiceJSONPath, messages, msg)
				}
			} else {
				postMessage(msg)
				messages := PostMessages{}
				writeJSON(*ServiceJSONPath, messages, msg)
			}
		} else {
			if isJSON(*ServiceJSONPath) {
				os.Remove(*ServiceJSONPath)
				postMessage(Message{Text: "遅延が解消されました。"})
			}
		}
	}
}
