package main

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

const WEBHOOK_ID = "WEBHOOK_ID"
const WEBHOOK_TOKEN = "WEBHOOK_TOKEN"

var discordClient *discordgo.Session

func init() {
	discordClient, _ = discordgo.New("")
}

func sendAccountError(account string, requestsSinceError int) {
	webhook := discordgo.WebhookParams{
		Content: account + ": Unexpected rate-limit, retrying in 1 hour. Requests since last error: " + strconv.Itoa(requestsSinceError),
	}

	discordClient.WebhookExecute(WEBHOOK_ID, WEBHOOK_TOKEN, false, &webhook)
}

func sendAsinError(account string, asin []string) {
	webhook := discordgo.WebhookParams{
		Content: account + ": Unrecognized ASIN requested " + strings.Join(asin, ","),
	}

	discordClient.WebhookExecute(WEBHOOK_ID, WEBHOOK_TOKEN, false, &webhook)
}
