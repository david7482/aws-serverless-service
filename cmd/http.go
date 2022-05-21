package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/app"
	"github.com/david7482/aws-serverless-service/internal/router"
)

func runHTTPServer(rootCtx context.Context, wg *sync.WaitGroup, port int, app *app.Application) {
	// Set to release mode to disable Gin logger
	gin.SetMode(gin.ReleaseMode)

	// Create gin router
	ginRouter := gin.New()
	setMiddlewares(rootCtx, ginRouter)

	// Register all handlers
	router.RegisterHandlers(ginRouter, app)

	// Build HTTP server
	httpAddr := fmt.Sprintf("0.0.0.0:%d", port)
	server := &http.Server{
		Addr:    httpAddr,
		Handler: ginRouter,
	}

	// Run the server in a goroutine
	go func() {
		zerolog.Ctx(rootCtx).Info().Msgf("HTTP server is on http://%s", httpAddr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			zerolog.Ctx(rootCtx).Panic().Err(err).Str("addr", httpAddr).Msg("fail to start HTTP server")
		}
	}()

	// Wait for rootCtx done
	go func() {
		<-rootCtx.Done()

		// Graceful shutdown http server with a timeout
		zerolog.Ctx(rootCtx).Info().Msgf("HTTP server is closing")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("fail to shutdown HTTP server")
		}

		// Notify when server is closed
		zerolog.Ctx(rootCtx).Info().Msgf("HTTP server is closed")
		wg.Done()
	}()
}

func setMiddlewares(ctx context.Context, ginRouter *gin.Engine) {
	ginRouter.Use(gin.Recovery())
	ginRouter.Use(requestid.New())
	ginRouter.Use(loggerMiddleware(ctx))
}

// This logger is referenced from gin's logger implementation with additional capabilities:
// 1. use zerolog to do structure log
// 2. add requestID into context logger
func loggerMiddleware(rootCtx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Ignore health-check to avoid polluting API logs
		if path == "/api/v1/health" {
			c.Next()
			return
		}

		// Add RequestID into the logger of the request context
		requestID := requestid.Get(c)
		zlog := zerolog.Ctx(rootCtx).With().Str("requestID", requestID).Str("path", c.FullPath()).Logger()
		c.Request = c.Request.WithContext(zlog.WithContext(rootCtx))

		// Process request
		c.Next()

		// Build all information that we want to log
		end := time.Now()
		params := gin.LogFormatterParams{
			TimeStamp:    end,
			Latency:      end.Sub(start),
			ClientIP:     c.ClientIP(),
			Method:       c.Request.Method,
			StatusCode:   c.Writer.Status(),
			ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
			BodySize:     c.Writer.Size(),
		}
		if raw != "" {
			path = path + "?" + raw
		}
		params.Path = path

		l := zerolog.Ctx(c.Request.Context()).Info().
			Time("callTime", params.TimeStamp).
			Str("requestID", requestid.Get(c)).
			Int("status", params.StatusCode).
			Dur("latency", params.Latency).
			Str("clientIP", params.ClientIP).
			Str("method", params.Method).
			Str("fullPath", params.Path).
			Str("userAgent", c.Request.Header.Get("User-Agent"))
		if params.ErrorMessage != "" {
			l = l.Err(errors.New(params.ErrorMessage))
		}
		l.Send()
	}
}
