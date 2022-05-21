package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/david7482/aws-serverless-service/internal/app"
	"github.com/david7482/aws-serverless-service/internal/domain"
)

func CreateLineChannel(app *app.Application) gin.HandlerFunc {
	type Body struct {
		ExternalChannelID     string `json:"externalChannelID" binding:"required"`
		ExternalChannelSecret string `json:"externalChannelSecret" binding:"required"`
	}

	type Response struct {
		ID                    int       `json:"id"`
		Name                  string    `json:"name"`
		ExternalChannelID     string    `json:"externalChannelID"`
		ExternalChannelSecret string    `json:"externalChannelSecret"`
		CreatedAt             time.Time `json:"created_at"`
	}
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var body Body
		err := c.ShouldBind(&body)
		if err != nil {
			respondWithError(c, domain.NewParameterError("invalid parameter", err))
			return
		}

		channel, err := app.ChannelService.CreateChannel(ctx, body.ExternalChannelID, body.ExternalChannelSecret)
		if err != nil {
			respondWithError(c, err)
			return
		}

		res := Response{
			ID:                    channel.ID,
			Name:                  channel.Name,
			ExternalChannelID:     channel.ExternalChannelID,
			ExternalChannelSecret: channel.ExternalChannelSecret,
			CreatedAt:             channel.CreatedAt,
		}

		respondWithJSON(c, http.StatusCreated, res)
		return
	}
}
