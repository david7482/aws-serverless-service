package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/app"
	"github.com/david7482/aws-serverless-service/internal/domain"
)

func ReceiveWebhookFromLine(app *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// We would always return 200 for LINE webhook
		defer func() {
			respondWithoutBody(c, http.StatusOK)
		}()

		// Validate parameters
		externalChannelID := c.Param("external_channel_id")
		if externalChannelID == "" {
			zerolog.Ctx(ctx).Error().Msg("no external channel ID")
			return
		}
		lineSignature := c.GetHeader("X-Line-Signature")
		if lineSignature == "" {
			zerolog.Ctx(ctx).Error().Str("externalChannelID", externalChannelID).Msg("no line signature")
			return
		}

		payload, err := c.GetRawData()
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("externalChannelID", externalChannelID).Msg("failed to read request body")
			return
		}
		if len(payload) == 0 {
			zerolog.Ctx(ctx).Error().Str("externalChannelID", externalChannelID).Msg("empty request body")
			return
		}

		// Invoke service
		err = app.MsgService.ReceiveWebhookFromLine(ctx, domain.LineWebhook{
			ExternalChannelID: externalChannelID,
			Signature:         lineSignature,
			Payload:           payload,
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to process LINE webhook events")
		}
	}
}
