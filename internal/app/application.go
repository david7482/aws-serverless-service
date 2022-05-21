package app

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/david7482/aws-serverless-service/internal/adapter/line"
	"github.com/david7482/aws-serverless-service/internal/adapter/postgres"
	"github.com/david7482/aws-serverless-service/internal/app/service/channel"
	"github.com/david7482/aws-serverless-service/internal/app/service/chat"
)

type Application struct {
	Params         ApplicationParams
	ChatService    *chat.ChatService
	ChannelService *channel.ChannelService

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

	lineService := line.NewLineService(ctx)

	app := &Application{
		Params: params,
		ChatService: chat.NewChatService(ctx, chat.ChatServiceParam{
			ChannelRepo: postgresRepo,
			LineService: lineService,
		}),
		ChannelService: channel.NewChannelService(ctx, channel.ChannelServiceParam{
			ChannelRepo: postgresRepo,
			LineService: lineService,
		}),
	}

	return app, nil
}
