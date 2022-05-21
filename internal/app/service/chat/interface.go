package chat

import (
	"context"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

//go:generate mockgen -destination automock/channel_repository.go -package=automock . ChannelRepository
type ChannelRepository interface {
	GetChannelByExternalID(ctx context.Context, externalChannelID string) (*domain.Channel, domain.Error)
	GetChannelByID(ctx context.Context, channelID int) (*domain.Channel, domain.Error)
}

//go:generate mockgen -destination automock/line_service.go -package=automock . LineService
type LineService interface {
	ValidateSignature(ctx context.Context, externalChannelSecret, signature string, payload []byte) bool
	ParseLineEvents(ctx context.Context, payload []byte) ([]domain.LineEvent, domain.Error)
}
