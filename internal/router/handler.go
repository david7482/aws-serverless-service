package router

import (
	"html/template"

	"github.com/gin-gonic/gin"

	"github.com/david7482/aws-serverless-service/html"
	"github.com/david7482/aws-serverless-service/internal/app"
)

func RegisterHandlers(router *gin.Engine, app *app.Application) {
	registerAPIHandlers(router, app)
	registerHTMLHandlers(router, app)
}

func registerAPIHandlers(router *gin.Engine, app *app.Application) {
	// We mount all handlers under /api path
	r := router.Group("/api")
	v1 := r.Group("/v1")

	// Add health-check
	v1.GET("/health", handlerHealthCheck())

	// Add webhook namespace
	webhookGroup := v1.Group("/webhook")
	{
		webhookGroup.POST("/line/:external_channel_id/events", ReceiveWebhookFromLine(app))
	}

	// Add channel namespace
	channelGroup := v1.Group("/channel")
	{
		channelGroup.POST("/line/channels", CreateLineChannel(app))
	}
}

func registerHTMLHandlers(router *gin.Engine, app *app.Application) {
	// Loads HTML files
	templ := template.Must(template.New("").ParseFS(html.F, "templates/*.html"))
	router.SetHTMLTemplate(templ)

	router.GET("/slide", RenderSlidePage(app))
}
