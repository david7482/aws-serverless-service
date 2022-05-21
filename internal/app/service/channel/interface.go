package channel

import (
	"context"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

//go:generate mockgen -destination automock/channel_repository.go -package=automock . ChannelRepository
type ChannelRepository interface {
	CreateChannel(ctx context.Context, channel domain.Channel) (*domain.Channel, domain.Error)
}

//go:generate mockgen -destination automock/line_service.go -package=automock . LineService
type LineService interface {
	IssueAccessToken(ctx context.Context, ExternalChannelID string, ExternalChannelSecret string) (string, time.Time, domain.Error)
	GetChannelInfo(ctx context.Context, accessToken string) (*linebot.BotInfoResponse, domain.Error)
}
