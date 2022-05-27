package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/adapter/line"
	"github.com/david7482/aws-serverless-service/internal/adapter/postgres"
)

var rootLogger zerolog.Logger
var pgRepo *postgres.PostgresRepository

const channelID = 1

func main() {
	const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"
	zerolog.TimeFieldFormat = rfc3339Milli
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	rootLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Create repositories
	databaseDSN := os.Getenv("DATABASE_DSN")
	db := sqlx.MustOpen("postgres", databaseDSN)
	if err := db.Ping(); err != nil {
		rootLogger.Error().Err(err).Str("databaseDSN", databaseDSN).Msg("fail to connect to database")
		return
	}
	pgRepo = postgres.NewPostgresRepository(context.Background(), db)

	lambda.Start(handler)
}

type event struct {
	DetailType string `json:"detail-type"`
	Source     string `json:"source"`
	Detail     struct {
		ChannelAccessToken string        `json:"channelAccessToken"`
		ExternalMemberID   string        `json:"externalMemberID"`
		EventType          string        `json:"eventType"`
		ReplyToken         string        `json:"replyToken"`
		EventContent       linebot.Event `json:"eventContent"`
	} `json:"detail"`
}

func isDownloadSlideMsg(e linebot.Event) bool {
	switch message := e.Message.(type) {
	case *linebot.TextMessage:
		if message.Text == "Download Slide" {
			return true
		}
	}
	return false
}

func handler(ctx context.Context, msg json.RawMessage) error {
	var logger zerolog.Logger
	lambdaCtx, ok := lambdacontext.FromContext(ctx)
	if ok {
		logger = rootLogger.With().Str("requestID", lambdaCtx.AwsRequestID).Logger()
	}
	logger.Info().RawJSON("msg", msg).Msg("raw message")

	var e event
	err := json.Unmarshal(msg, &e)
	if err != nil {
		logger.Error().Err(err).Msg("fail to unmarshal msg to event")
		return err
	}

	if !isDownloadSlideMsg(e.Detail.EventContent) {
		// do nothing if it's not DownloadSlide message
		return nil
	}

	// Query necessary information from repository
	url, err := pgRepo.GetEnabledSlideURL(ctx, channelID)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("fail to get enabled slide URL")
		return err
	}

	// Reply the image message with slide URL
	lineSrv := line.NewLineService(ctx)
	err = lineSrv.SendMessage(ctx, line.SendMessageParams{
		AccessToken: e.Detail.ChannelAccessToken,
		ReplyToken:  e.Detail.ReplyToken,
		To:          e.Detail.ExternalMemberID,
		Messages:    []linebot.SendingMessage{linebot.NewImageMessage(url, url)},
	})
	if err != nil {
		logger.Error().Err(err).Msg("fail to unmarshal msg to event")
		return err
	}

	return nil
}
