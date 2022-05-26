package eventbridge

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

const (
	eventBridgeSource     = "chatbot"
	eventBridgeDetailType = "line-message"
)

type EventBridge struct {
	evb     *eventbridge.EventBridge
	busName string
}

func NewEventBridge(_ context.Context, s *session.Session, busName string) *EventBridge {
	evb := eventbridge.New(s)
	return &EventBridge{
		evb:     evb,
		busName: busName,
	}
}

func (e *EventBridge) PutEvent(ctx context.Context, data string) domain.Error {
	entries := []*eventbridge.PutEventsRequestEntry{{
		Source:       aws.String(eventBridgeSource),
		DetailType:   aws.String(eventBridgeDetailType),
		Detail:       aws.String(data),
		EventBusName: aws.String(e.busName),
	}}

	_, err := e.evb.PutEventsWithContext(ctx, &eventbridge.PutEventsInput{Entries: entries})
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("fail to put event to eventbridge")
		return domain.NewExternalError("", nil, err)
	}

	return nil
}
