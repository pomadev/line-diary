package main

import (
	"fmt"
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
	Date   string

	Content string `dynamo:"Content"`
}

// TODO: エラー返す処理重複を除く
// TODO: ネスト地獄解消
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
		if e.Type == linebot.EventTypePostback {
			if e.Postback.Params.Date != "" {
				c, err := fetch(e.Source.UserID, e.Postback.Params.Date)
				if err != nil {
					if err != dynamo.ErrNotFound {
						log.Printf("cannot fetch: %v", err)
						return events.APIGatewayProxyResponse{
							StatusCode: http.StatusInternalServerError,
						}, nil
					}
					if _, err = bot.ReplyMessage(e.ReplyToken, linebot.NewTextMessage("日記が見つかりませんでした")).Do(); err != nil {
						log.Printf("bot cannot reply message: %v", err)
					}
				} else {
					if _, err = bot.ReplyMessage(e.ReplyToken, linebot.NewTextMessage(c.Content)).Do(); err != nil {
						log.Printf("bot cannot reply message: %v", err)
					}
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

// TODO: dynamodb処理処理共通化
func put(id, content string, t time.Time) error {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(os.Getenv("DYNAMODB_REGION"))})
	table := db.Table(os.Getenv("DYNAMODB_TABLE"))

	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60))

	d := diary{
		LineID:  id,
		Date:    fmt.Sprintf("%d/%d/%d", t.Year(), t.Month(), t.Day()),
		Content: content,
	}
	err := table.Put(d).Run()
	return err
}

func fetch(id, s string) (diary, error) {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(os.Getenv("DYNAMODB_REGION"))})
	table := db.Table(os.Getenv("DYNAMODB_TABLE"))

	// TODO: 処理切り分け
	layout := "2006-01-02"
	t, err := time.Parse(layout, s)
	if err != nil {
		return diary{}, err
	}

	var d diary
	err = table.Get("LineID", id).
		Range("Date", dynamo.Equal, fmt.Sprintf("%d/%d/%d", t.Year(), t.Month(), t.Day())).
		One(&d)
	return d, err
}

func main() {
	lambda.Start(handler)
}
