package slide

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

type Repository interface {
	GetSlideURLByPage(ctx context.Context, channelID, page int) (string, int, domain.Error)
	GetLastPageNumber(ctx context.Context, channelID int) (int, domain.Error)
	UpdateCurrentPage(ctx context.Context, channelID, page int) domain.Error
}

type SlideService struct {
	repo Repository
}

func NewSlideService(_ context.Context, repo Repository) *SlideService {
	return &SlideService{
		repo: repo,
	}
}

// GetSlideURL returns the slide URL specified by page. If p is out of scope,
// it would return page 1.
func (s *SlideService) GetSlideURL(ctx context.Context, channelID, p int) (url string, page int, err domain.Error) {
	last, err := s.repo.GetLastPageNumber(ctx, channelID)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("fail to get last page")
		return "", 0, err
	}

	if p > last {
		// if p > last, we jump back to the 1st page
		p = 1
	}

	url, page, err = s.repo.GetSlideURLByPage(ctx, channelID, p)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Int("page", p).Msg("fail to get slide url")
		return "", 0, err
	}

	return url, page, nil
}

// GetPrevNext returns the previous and next page number
func (s *SlideService) GetPrevNext(ctx context.Context, channelID, p int) (prev int, next int, err domain.Error) {
	last, err := s.repo.GetLastPageNumber(ctx, channelID)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Int("page", p).Msg("fail to get last page")
		return 0, 0, err
	}

	if p == 1 {
		prev = last
	} else {
		prev = p - 1
	}

	if p == last {
		next = 1
	} else {
		next = p + 1
	}

	return prev, next, err
}

// UpdateCurrentPage marked the specified page to be enabled, and all other pages are disabled.
func (s *SlideService) UpdateCurrentPage(ctx context.Context, channelID, p int) domain.Error {
	err := s.repo.UpdateCurrentPage(ctx, channelID, p)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Int("page", p).Msg("fail to update current page")
		return err
	}
	return nil
}
