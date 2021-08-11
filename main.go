package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
)

func writeJson(name string, objs ...interface{}) {
	res, err := json.MarshalIndent(objs, "", "  ")
	if err != nil {
		log.Fatalln("failed to serialize json ", name, ":", err)
		return
	}
	name += ".json"
	err = os.WriteFile(name, res, 0600)
	if err != nil {
		log.Fatalln("failed to write data to ", name, ":", err)
		return
	}
}

func main() {

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Fatalln("Envvar SLACK_TOKEN is not set")
	}

	api := slack.New(token, slack.OptionDebug(false))

	//api.GetEmoji()

	fmt.Print("Fetching users...")
	users, err := api.GetUsers()
	if err != nil {
		log.Fatalln("failed:", err)
		return
	}
	fmt.Println("got ", len(users))
	writeJson("users", users)

	fmt.Print("Fetching conversations...")
	channels, _, err := api.GetConversations(&slack.GetConversationsParameters{Limit: 500, Types: []string{"public_channel", "private_channel", "mpim", "im"}})
	if err != nil {
		log.Fatalln("failed:", err)
		return
	}

	fmt.Println("got ", len(channels))

	channelPathBase := "channels"
	if err != nil {
		log.Fatalln("failed to create channel folder:", err)
		return
	}
	for _, channel := range channels {
		channelPath := channelPathBase + "/" + channel.NameNormalized + "-" + channel.ID
		err = os.MkdirAll(channelPath, 0755)
		writeJson(channelPath+"/info", channel)

		fmt.Printf("  Fetching %s...", channel.Name)
		var messages []slack.Message
		cursor := ""
		for {

			resp, err := api.GetConversationHistory(&slack.GetConversationHistoryParameters{ChannelID: channel.ID, Limit: 500, Cursor: cursor})
			if err != nil {
				log.Fatal(err)
				return
			}
			messages = append(messages, resp.Messages...)
			if resp.HasMore {
				fmt.Printf("got %d [last ts %s].. ", len(messages), messages[len(messages)-1].Timestamp)
				cursor = resp.ResponseMetaData.NextCursor
			} else {
				break
			}
		}
		fmt.Println("got ", len(messages))

		writeJson(channelPath+"/messages", messages)
	}
	/*
		fmt.Print("Fetching files...")
		files, _, err := api.GetFiles(slack.GetFilesParameters{})
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, file := range files {
			fmt.Printf("File ID: %s, Name: %s :: %+v\n", file.ID, file.Name, file)
		}
	*/
}
