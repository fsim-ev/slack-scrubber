# Slack Scrubber

Uses Slack API to fetch importannt data (messages, files) from accessable channels, groups und private chats.
Before execution you need to:

 - create a Slack App with the right permissions (`*.list; *.read`)
 - install it into your desired Workspace
 - placing the app auth token into `SLACK_TOKEN` environment variable

 Execute with `go run main.go`.
