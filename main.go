package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

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

	fmt.Print("Fetching users... ")
	users, err := api.GetUsers()
	if err != nil {
		log.Fatalln("failed:", err)
		return
	}
	fmt.Println("got", len(users))
	writeJson("users", users)

	fmt.Print("Fetching conversations... ")
	channels, _, err := api.GetConversations(&slack.GetConversationsParameters{Limit: 500, Types: []string{"public_channel", "private_channel", "mpim", "im"}})
	if err != nil {
		log.Fatalln("failed:", err)
		return
	}
	fmt.Println("got", len(channels))

	channelPathBase := "channels"
	if err != nil {
		log.Fatalln("failed to create channel folder:", err)
		return
	}
	for _, channel := range channels {
		channelName := channel.Name
		if channelName == "" && channel.IsIM {
			channelName = "im-" + channel.User
		}

		channelPath := channelPathBase + "/" + channelName + "-" + channel.ID
		err = os.MkdirAll(channelPath, 0755)
		writeJson(channelPath+"/info", channel)

		fmt.Printf("  Fetching %s... ", channelName)
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
		fmt.Println("got", len(messages))
		if len(messages) > 0 {
			writeJson(channelPath+"/messages", messages)
		}
	}

	fmt.Print("Fetching files... ")
	var files []slack.File
	page_num := 0
	for {
		files_chunk, page, err := api.GetFiles(slack.GetFilesParameters{Count: 500, ShowHidden: true, Page: page_num})
		if err != nil {
			log.Fatal(err)
			return
		}

		files = append(files, files_chunk...)
		fmt.Printf("got %d [last ts %d].. ", len(files), files[len(files)-1].Timestamp.Time().Unix())
		if page.Page != page.Pages {
			page_num += 1
		} else {
			break
		}
	}
	fmt.Println("got", len(files))
	writeJson("files", files)

	fmt.Println("Downloading files... ")
	filePathBase := "files"
	err = os.MkdirAll(filePathBase, 0755)
	if err != nil {
		log.Fatalln("failed to create files folder:", err)
		return
	}
	for _, file_meta := range files {
		fmt.Print("  ", file_meta.Name, " ... ")

		filePath := filePathBase + "/" + file_meta.Name
		// check for existing
		file_stat, err := os.Stat(filePath)
		if err == nil {
			if file_stat.Size() == int64(file_meta.Size) {
				fmt.Println(" same file exists (skipping)")
				continue
			} else {
				fileName := file_meta.Name[0:len(file_meta.Name)-len(file_meta.Filetype)-1] + "-" + file_meta.ID + "." + file_meta.Filetype
				filePath = filePathBase + "/" + fileName
				fmt.Print(" -> ", fileName, " ... ")

				file_stat, err = os.Stat(filePath)
				if err == nil && file_stat.Size() == int64(file_meta.Size) {
					fmt.Println(" same file exists (skipping)")
					continue
				}
			}
		}

		file, err := os.Create(filePath)
		if err != nil {
			fmt.Println("failed to create file ", file_meta.Name, " :", err)
			continue
		}

		err = api.GetFile(file_meta.URLPrivateDownload, file)
		if err != nil {
			fmt.Println("failed to download file ", file_meta.Name, " :", err)
			continue
		}

		file.Close()
		fmt.Println("done")
	}

	fmt.Println("Fetching emojis... ")
	emojis, err := api.GetEmoji()
	if err != nil {
		log.Fatalln("failed to fetch emojis:", err)
		return
	}
	fmt.Println("got", len(emojis))

	fmt.Println("Downloading emojis... ")
	emojiPathBase := "emojis"
	err = os.MkdirAll(emojiPathBase, 0755)
	if err != nil {
		log.Fatalln("failed to create emoji folder:", err)
		return
	}
	for emoji, url := range emojis {
		fmt.Print("  ", emoji, " ... ")

		filePath := emojiPathBase + "/" + emoji + "." + path.Ext(url)
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Println("failed to create emoji file ", emoji, " :", err)
			continue
		}

		err = api.GetFile(url, file)
		if err != nil {
			fmt.Println("failed to download emoji file ", emoji, " :", err)
			continue
		}
		fmt.Println("done")
	}
}
