package chat

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

type ChatService struct {
	channelRepo ChannelRepository
	lineService LineService
}

type ChatServiceParam struct {
	ChannelRepo ChannelRepository
	LineService LineService
}

func NewChatService(_ context.Context, param ChatServiceParam) *ChatService {
	return &ChatService{
		channelRepo: param.ChannelRepo,
		lineService: param.LineService,
	}
}

// logger wrap the execution context with component info
func (s *ChatService) logger(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx).With().Str("service", "chat").Logger()
	return &l
}

func (s *ChatService) ReceiveWebhookFromLine(ctx context.Context, webhook domain.LineWebhook) domain.Error {
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

	// Handle Line events, we support Message, Follow, Unfollow events so far.
	for _, event := range events {
		s.logger(ctx).Info().
			Str("externalMemberID", event.ExternalMemberID).
			Bytes("eventContent", event.EventContent).
			Msg("get line event")
	}
	return nil
}

func (s *ChatService) decodeLineWebhook(ctx context.Context, externalChannelSecret string, webhook domain.LineWebhook) ([]domain.LineEvent, domain.Error) {
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
