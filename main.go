package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	// SlackURL は Slack チャンネルの URL
	SlackURL = os.Getenv("SLACK_URL")
	// ServiceJSONPath は同じ内容の遅延情報を送信しないために使用する json ファイルのパス
	ServiceJSONPath = "./json/service.json"
)

// Message は送信するメッセージを保持する構造体
type Message struct {
	Text string `json:"text"`
}

// PostMessages は送信するメッセージ群を保持する構造体
type PostMessages struct {
	Messages []Message `json:"messages"`
}

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

func isJSON() bool {
	if _, err := os.Stat(ServiceJSONPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func readJSON() PostMessages {
	var messages PostMessages
	raw, _ := ioutil.ReadFile(ServiceJSONPath)
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

func writeJSON(msgs PostMessages, msg Message) {
	msgs.Messages = append(msgs.Messages, msg)
	jsonByte, err := json.Marshal(msgs)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	ioutil.WriteFile(ServiceJSONPath, jsonByte, os.ModePerm)
}

func main() {
	msgs := makeMessages()
	for _, msg := range msgs.Messages {
		// if strings.Index(msg.Text, "発生しています") > -1 {
		if strings.Index(msg.Text, "発生しています") != -1 {
			if isJSON() {
				messages := readJSON()
				if !isExistMsg(msg, messages) {
					postMessage(msg)
					writeJSON(messages, msg)
				}
			} else {
				postMessage(msg)
				messages := PostMessages{}
				writeJSON(messages, msg)
			}
		} else {
			if isJSON() {
				os.Remove(ServiceJSONPath)
				postMessage(Message{Text: "遅延が解消されました。"})
			}
		}
	}
}
