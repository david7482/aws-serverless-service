package channel

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

type ChannelService struct {
	channelRepo ChannelRepository
	lineService LineService
}

type ChannelServiceParam struct {
	ChannelRepo ChannelRepository
	LineService LineService
}

func NewChannelService(_ context.Context, param ChannelServiceParam) *ChannelService {
	return &ChannelService{
		channelRepo: param.ChannelRepo,
		lineService: param.LineService,
	}
}

// logger wrap the execution context with component info
func (s *ChannelService) logger(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx).With().Str("service", "channel").Logger()
	return &l
}

func (s *ChannelService) CreateChannel(ctx context.Context, externalChannelID, externalChannelSecret string) (*domain.Channel, domain.Error) {
	accessToken, expiredAt, err := s.lineService.IssueAccessToken(ctx, externalChannelID, externalChannelSecret)
	if err != nil {
		s.logger(ctx).Error().Err(err).Msg("failed to issue access token from Line")
		return nil, err
	}

	info, err := s.lineService.GetChannelInfo(ctx, accessToken)
	if err != nil {
		s.logger(ctx).Error().Err(err).Msg("failed to get channel info")
		return nil, err
	}

	channel := &domain.Channel{
		Name:                  info.DisplayName,
		ExternalChannelID:     externalChannelID,
		ExternalChannelSecret: externalChannelSecret,
		AccessToken:           accessToken,
		AccessTokenExpiredAt:  expiredAt,
	}

	channel, err = s.channelRepo.CreateChannel(ctx, *channel)
	if err != nil {
		s.logger(ctx).Error().Err(err).Msg("failed to create channel")
		return nil, err
	}
	return channel, nil
}
