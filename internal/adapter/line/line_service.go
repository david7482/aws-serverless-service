package line

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

type LineService struct {
	client *resty.Client
}

func NewLineService(_ context.Context) *LineService {
	client := resty.New().
		SetTimeout(15*time.Second).
		SetRetryCount(2).
		SetHeader("user-agent", "chatbot/1")
	return &LineService{
		client: client,
	}
}

func (s *LineService) logger(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx).With().Str("service", "line").Logger()
	return &l
}

type issueAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (s *LineService) IssueAccessToken(ctx context.Context, ExternalChannelID string, ExternalChannelSecret string) (string, time.Time, domain.Error) {
	// Reference: https://developers.line.biz/en/reference/messaging-api/#revoke-channel-access-token-v2-1

	uri := "https://api.line.me/v2/oauth/accessToken"
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("client_id", ExternalChannelID)
	form.Add("client_secret", ExternalChannelSecret)
	body := strings.NewReader(form.Encode())

	var ret issueAccessTokenResponse
	resp, err := s.client.R().
		SetHeader("content-type", "application/x-www-form-urlencoded").
		SetBody(body).
		SetResult(&ret).
		Post(uri)

	if err != nil {
		return "", time.Now(), domain.NewExternalError("", nil, err)
	}
	if !resp.IsSuccess() {
		code := resp.StatusCode()
		s.logger(ctx).Error().
			Int("statusCode", code).
			Msg("failed to issue access token")
		return "", time.Now(), domain.NewExternalError("", &code, errors.New("failed to issue access token"))
	}
	return ret.AccessToken, time.Now().Add(time.Duration(ret.ExpiresIn) * time.Second), nil
}

func (s *LineService) ValidateSignature(_ context.Context, externalChannelSecret, signature string, payload []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	hash := hmac.New(sha256.New, []byte(externalChannelSecret))

	_, err = hash.Write(payload)
	if err != nil {
		return false
	}

	return hmac.Equal(decoded, hash.Sum(nil))
}

func (s *LineService) ParseLineEvents(_ context.Context, payload []byte) ([]domain.LineEvent, domain.Error) {
	request := struct {
		LineEvents []linebot.Event `json:"events"`
	}{}
	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, domain.NewInternalError("", err)
	}
	var events []domain.LineEvent
	for _, lineEvent := range request.LineEvents {
		content, _ := lineEvent.MarshalJSON()
		event := domain.LineEvent{
			EventType:        domain.LineEventType(lineEvent.Type),
			ExternalMemberID: lineEvent.Source.UserID,
			ReplyToken:       lineEvent.ReplyToken,
			EventContent:     content,
		}

		events = append(events, event)
	}

	return events, nil
}

func (s *LineService) GetChannelInfo(ctx context.Context, accessToken string) (*linebot.BotInfoResponse, domain.Error) {
	bot, err := linebot.New("no use", accessToken)
	if err != nil {
		return nil, domain.NewInternalError("", err)
	}

	info, err := bot.GetBotInfo().WithContext(ctx).Do()
	if err != nil {
		return nil, domain.NewExternalError("", nil, err)
	}
	return info, nil
}

type SendMessageParams struct {
	AccessToken string
	Messages    []linebot.SendingMessage
	ReplyToken  string
	To          string // userId, groupId, roomId
	RetryKey    string
}

// SendMessage would take care of PushMessage and ReplyMessage internally. It would also fall back
// to PushMessage if ReplyMessage fail.
func (s *LineService) SendMessage(ctx context.Context, params SendMessageParams) (err error) {
	bot, _ := linebot.New("not-used", params.AccessToken)

	// If we have reply token, we would try ReplyMessage() first.
	if params.ReplyToken != "" {
		_, err = bot.ReplyMessage(params.ReplyToken, params.Messages...).
			WithContext(ctx).
			Do()
		if err == nil {
			return nil
		}
	}

	// Try pushMessage() if we have To field
	if params.To != "" {
		_, err = bot.PushMessage(params.To, params.Messages...).
			WithContext(ctx).
			WithRetryKey(params.RetryKey).
			Do()
		if err == nil {
			return nil
		}
	}

	return err
}
