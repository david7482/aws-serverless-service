package app

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/david7482/aws-serverless-service/internal/adapter/eventbridge"
	"github.com/david7482/aws-serverless-service/internal/adapter/line"
	"github.com/david7482/aws-serverless-service/internal/adapter/postgres"
	"github.com/david7482/aws-serverless-service/internal/app/service/channel"
	"github.com/david7482/aws-serverless-service/internal/app/service/message"
	"github.com/david7482/aws-serverless-service/internal/app/service/slide"
)

type Application struct {
	Params         ApplicationParams
	MsgService     *message.MessageService
	ChannelService *channel.ChannelService
	SlideService   *slide.SlideService

	//AccountService          *auth.AccountService
	//TokenService            *auth.TokenService
	//OrgService              *organization.OrgService
	//UserService             *organization.UserService
	//RoleService             *organization.RoleService
	//WorkerTaskService       *workertask.WorkerTaskService
	//ChannelService          *organization.ChannelService
	//TagService              *tag.TagService
	//TagChannelMemberService *tag.TagChannelMemberService
}

type ApplicationParams struct {
	// Database parameters
	DatabaseDSN string

	// AWS parameters
	AWSRegion          string
	AWSEventBridgeName string
}

func MustNewApplication(ctx context.Context, params ApplicationParams) *Application {
	app, err := NewApplication(ctx, params)
	if err != nil {
		log.Panicf("fail to new application, err: %s", err.Error())
	}
	return app
}

func NewApplication(ctx context.Context, params ApplicationParams) (*Application, error) {
	// Create repositories
	db := sqlx.MustOpen("postgres", params.DatabaseDSN)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	postgresRepo := postgres.NewPostgresRepository(ctx, db)

	// Create AWS resources
	ses, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(params.AWSRegion),
		},
	})
	if err != nil {
		return nil, err
	}
	eventBridge := eventbridge.NewEventBridge(ctx, ses, params.AWSEventBridgeName)

	lineService := line.NewLineService(ctx)

	app := &Application{
		Params: params,
		MsgService: message.NewMessageService(ctx, message.MessageServiceParam{
			ChannelRepo: postgresRepo,
			LineService: lineService,
			EventBridge: eventBridge,
		}),
		ChannelService: channel.NewChannelService(ctx, channel.ChannelServiceParam{
			ChannelRepo: postgresRepo,
			LineService: lineService,
		}),
		SlideService: slide.NewSlideService(ctx, postgresRepo),
	}

	return app, nil
}
