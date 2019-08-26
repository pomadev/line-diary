package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/line/line-bot-sdk-go/linebot"
)

// TODO: 共通化
type diary struct {
	LineID string
	Date   string

	Content string `dynamo:"Content"`
}

func handler() {
	// TODO: 共通化
	bot, _ := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)

	// TODO: 共通化
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(os.Getenv("DYNAMODB_REGION"))})
	table := db.Table(os.Getenv("DYNAMODB_TABLE"))

	t := time.Now().AddDate(0, 0, -7)

	var d []diary
	err := table.Scan().Filter("'Date' = ?", fmt.Sprintf("%d/%d/%d", t.Year(), t.Month(), t.Day())).All(&d)
	if err != nil {
		log.Printf("cannot get all: %v", err)
	}

	for _, v := range d {
		msg := linebot.NewTextMessage(fmt.Sprintf("%vの日記です\n\n%v", v.Date, v.Content))
		_, err := bot.PushMessage(v.LineID, msg).Do()
		log.Printf("cannot push message: %v", err)
	}
}

func main() {
	lambda.Start(handler)
}
