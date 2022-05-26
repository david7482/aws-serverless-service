package router

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/david7482/aws-serverless-service/internal/app"
	"github.com/david7482/aws-serverless-service/internal/domain"
)

const channelID = 1

func RenderSlidePage(app *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get query parameter with default value
		page := c.DefaultQuery("page", "1")

		p, err := strconv.Atoi(page)
		if err != nil {
			respondWithError(c, domain.NewParameterError("", err))
			return
		}

		url, p, err := app.SlideService.GetSlideURL(ctx, channelID, p)
		if err != nil {
			respondWithError(c, err)
			return
		}

		prev, next, err := app.SlideService.GetPrevNext(ctx, channelID, p)
		if err != nil {
			respondWithError(c, err)
			return
		}

		err = app.SlideService.UpdateCurrentPage(ctx, channelID, p)
		if err != nil {
			respondWithError(c, err)
			return
		}

		// Call the HTML method of the Context to render a template
		c.HTML(http.StatusOK, "slide.html", gin.H{
			"img":  url,
			"prev": prev,
			"next": next,
		})
	}
}
