package message

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

type MessageService struct {
	channelRepo ChannelRepository
	lineService LineService
	eventBridge EventBridge
}

type MessageServiceParam struct {
	ChannelRepo ChannelRepository
	LineService LineService
	EventBridge EventBridge
}

func NewMessageService(_ context.Context, param MessageServiceParam) *MessageService {
	return &MessageService{
		channelRepo: param.ChannelRepo,
		lineService: param.LineService,
		eventBridge: param.EventBridge,
	}
}

// logger wrap the execution context with component info
func (s *MessageService) logger(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx).With().Str("service", "message").Logger()
	return &l
}

func (s *MessageService) ReceiveWebhookFromLine(ctx context.Context, webhook domain.LineWebhook) domain.Error {
	// Check if the given channel is existed
	channel, err := s.channelRepo.GetChannelByExternalID(ctx, webhook.ExternalChannelID)
	if err != nil {
		s.logger(ctx).Error().Err(err).Msg("failed to get channel")
		return err
	}

	// Decode Line Webhook
	events, err := s.decodeLineWebhook(ctx, channel.ExternalChannelSecret, webhook)
	if err != nil {
		s.logger(ctx).Error().Err(err).Msg("failed to decode line webhook")
		return err
	}

	type event struct {
		ChannelAccessToken string          `json:"channelAccessToken"`
		EventType          string          `json:"eventType"`
		ExternalMemberID   string          `json:"externalMemberID"`
		ReplyToken         string          `json:"replyToken"`
		EventContent       json.RawMessage `json:"eventContent"`
	}

	// Handle Line events, we support Message, Follow, Unfollow events so far.
	for _, e := range events {
		s.logger(ctx).Info().
			Str("eventType", string(e.EventType)).
			Str("externalMemberID", e.ExternalMemberID).
			Bytes("eventContent", e.EventContent).
			Msg("get line event")

		evt := event{
			ChannelAccessToken: channel.AccessToken,
			ExternalMemberID:   e.ExternalMemberID,
			EventContent:       e.EventContent,
			ReplyToken:         e.ReplyToken,
			EventType:          string(e.EventType),
		}

		data, err := json.Marshal(evt)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("fail to marshal event")
			continue
		}

		_ = s.eventBridge.PutEvent(ctx, string(data))
	}
	return nil
}

func (s *MessageService) decodeLineWebhook(ctx context.Context, externalChannelSecret string, webhook domain.LineWebhook) ([]domain.LineEvent, domain.Error) {
	// Validate the payload
	valid := s.lineService.ValidateSignature(ctx, externalChannelSecret, webhook.Signature, webhook.Payload)
	if !valid {
		msg := "invalid webhook payload"
		s.logger(ctx).Error().Msg(msg)
		return nil, domain.NewParameterError(msg, errors.New(msg))
	}

	// Parse line events
	events, err := s.lineService.ParseLineEvents(ctx, webhook.Payload)
	if err != nil {
		s.logger(ctx).Error().Err(err).Msg("failed to parse line webhook events")
		return nil, err
	}
	return events, nil
}
