package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/line/line-bot-sdk-go/linebot"
)

type diary struct {
	LineID string
	Date   time.Time

	Content string `dynamo:"Content"`
}

// TODO: エラー返す処理重複を除く
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Printf("cannot initialize linebot: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}
	r, err := translate(request)
	if err != nil {
		log.Printf("cannot translate from events.APIGatewayProxyRequest to *http.Request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}
	es, err := bot.ParseRequest(r)
	if err != nil {
		log.Printf("bot cannot parse request: %v", err)
		if err == linebot.ErrInvalidSignature {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}
	for _, e := range es {
		if e.Type == linebot.EventTypeMessage {
			switch m := e.Message.(type) {
			case *linebot.TextMessage:
				err = put(e.Source.UserID, m.Text, e.Timestamp)
				if err != nil {
					log.Printf("cannot put data: %v", err)
					if _, err = bot.ReplyMessage(e.ReplyToken, linebot.NewTextMessage("保存失敗")).Do(); err != nil {
						log.Printf("bot cannot reply message: %v", err)
					}
					return events.APIGatewayProxyResponse{
						StatusCode: http.StatusInternalServerError,
					}, nil
				}
				if _, err = bot.ReplyMessage(e.ReplyToken, linebot.NewTextMessage("保存完了")).Do(); err != nil {
					log.Printf("bot cannot reply message: %v", err)
				}
			}
		}
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

func translate(request events.APIGatewayProxyRequest) (*http.Request, error) {
	r, err := http.NewRequest(request.HTTPMethod, "", strings.NewReader(request.Body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("X-Line-Signature", request.Headers["X-Line-Signature"])
	return r, nil
}

func put(id, content string, t time.Time) error {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(os.Getenv("DYNAMODB_REGION"))})
	table := db.Table(os.Getenv("DYNAMODB_TABLE"))

	d := diary{
		LineID:  id,
		Date:    time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)),
		Content: content,
	}
	err := table.Put(d).Run()
	return err
}

func main() {
	lambda.Start(handler)
}
