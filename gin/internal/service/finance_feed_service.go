package service

import (
	"context"

	"gin/internal/domain/finance_feed"
	repopg "gin/internal/repository/postgres"
)

type FinanceFeedService struct {
	repository *repopg.FinanceFeedRepository
}

func NewFinanceFeedService(repository *repopg.FinanceFeedRepository) *FinanceFeedService {
	return &FinanceFeedService{repository: repository}
}

func (s *FinanceFeedService) FakeDeposits(ctx context.Context, limit int) (finance_feed.FeedResponse, error) {
	return s.buildResponse(ctx, "deposit", limit)
}

func (s *FinanceFeedService) FakeWithdraws(ctx context.Context, limit int) (finance_feed.FeedResponse, error) {
	return s.buildResponse(ctx, "withdraw", limit)
}

func (s *FinanceFeedService) buildResponse(ctx context.Context, channel string, limit int) (finance_feed.FeedResponse, error) {
	var (
		records []repopg.FakeFinanceFeedRecord
		err     error
	)

	switch channel {
	case "deposit":
		records, err = s.repository.ListFakeDeposits(ctx, limit)
	case "withdraw":
		records, err = s.repository.ListFakeWithdraws(ctx, limit)
	}
	if err != nil {
		return finance_feed.FeedResponse{}, err
	}

	items := make([]finance_feed.Item, 0, len(records))
	for _, record := range records {
		channelCopy := channel
		items = append(items, finance_feed.Item{
			ID:          record.ID,
			MaskedCode:  record.MaskedCode,
			MaskedPhone: record.MaskedPhone,
			StatusLabel: record.StatusLabel,
			CreatedAt:   formatDateTime(record.CreatedAt),
			Channel:     &channelCopy,
		})
	}

	return finance_feed.FeedResponse{
		Message: "Lấy feed giao dịch gần đây thành công",
		Items:   items,
	}, nil
}
