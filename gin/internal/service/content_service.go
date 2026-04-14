package service

import (
	"context"
	"errors"
	"strings"

	"gin/internal/domain/content"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"
	"gin/internal/support/message"
)

type ContentService struct {
	repository       *repopg.ContentRepository
	contentAssetBase string
}

func NewContentService(repository *repopg.ContentRepository, contentAssetBase string) *ContentService {
	return &ContentService{
		repository:       repository,
		contentAssetBase: strings.TrimRight(strings.TrimSpace(contentAssetBase), "/"),
	}
}

func (s *ContentService) Home(ctx context.Context) (content.HomeResponse, error) {
	bannerRecords, err := s.repository.ListActiveBanners(ctx, 6, clock.Now())
	if err != nil {
		return content.HomeResponse{}, err
	}
	newsRecords, err := s.repository.ListLatestNews(ctx, 8)
	if err != nil {
		return content.HomeResponse{}, err
	}

	banners := make([]content.BannerItem, 0, len(bannerRecords))
	for _, item := range bannerRecords {
		banners = append(banners, content.BannerItem{
			ID:       item.ID,
			Title:    firstNonEmptyStringPtr(item.Title),
			ImageURL: s.buildAssetURL(item.ImagePath),
			LinkURL:  firstNonEmptyStringPtr(item.LinkURL),
		})
	}

	highlights := make([]content.NewsItem, 0, len(newsRecords))
	for _, item := range newsRecords {
		highlights = append(highlights, s.toNewsItem(item, false))
	}

	return content.HomeResponse{
		Message:    message.ContentHomeSuccess,
		Banners:    banners,
		Highlights: highlights,
	}, nil
}

func (s *ContentService) Promotions(ctx context.Context, page, pageSize int) (content.PromotionListResponse, error) {
	records, total, err := s.repository.ListPublishedNews(ctx, page, pageSize, true)
	if err != nil {
		return content.PromotionListResponse{}, err
	}

	items := make([]content.NewsItem, 0, len(records))
	for _, item := range records {
		items = append(items, s.toNewsItem(item, true))
	}

	return content.PromotionListResponse{
		Message:    message.ContentPromotionSuccess,
		Page:       normalizePage(page),
		PageSize:   normalizePageSize(pageSize),
		Total:      total,
		TotalPages: calcNotificationTotalPages(total, normalizePageSize(pageSize)),
		Items:      items,
	}, nil
}

func (s *ContentService) News(ctx context.Context, page, pageSize int) (content.NewsListResponse, error) {
	records, total, err := s.repository.ListPublishedNews(ctx, page, pageSize, false)
	if err != nil {
		return content.NewsListResponse{}, err
	}

	items := make([]content.NewsItem, 0, len(records))
	for _, item := range records {
		items = append(items, s.toNewsItem(item, true))
	}

	return content.NewsListResponse{
		Message:    message.ContentNewsSuccess,
		Page:       normalizePage(page),
		PageSize:   normalizePageSize(pageSize),
		Total:      total,
		TotalPages: calcNotificationTotalPages(total, normalizePageSize(pageSize)),
		Items:      items,
	}, nil
}

func (s *ContentService) NewsDetail(ctx context.Context, slug string) (content.NewsDetailResponse, error) {
	record, err := s.repository.FindPublishedNewsBySlug(ctx, slug)
	if err != nil {
		return content.NewsDetailResponse{}, err
	}
	relatedRecords, err := s.repository.ListRelatedNews(ctx, slug, 3)
	if err != nil {
		return content.NewsDetailResponse{}, err
	}

	related := make([]content.NewsItem, 0, len(relatedRecords))
	for _, item := range relatedRecords {
		related = append(related, s.toNewsItem(item, false))
	}

	return content.NewsDetailResponse{
		Message: message.ContentNewsDetailSuccess,
		Item:    s.toNewsItem(record, true),
		Related: related,
	}, nil
}

func (s *ContentService) toNewsItem(record repopg.NewsRecord, includeContent bool) content.NewsItem {
	item := content.NewsItem{
		ID:          record.ID,
		Title:       record.Title,
		Slug:        record.Slug,
		Excerpt:     firstNonEmptyStringPtr(record.Excerpt),
		CoverImage:  s.buildAssetURL(firstNonEmptyStringPtr(record.CoverImage)),
		PublishedAt: record.PublishedAt,
		CreatedAt:   record.CreatedAt,
	}
	if includeContent {
		item.Content = record.Content
	}
	return item
}

func (s *ContentService) buildAssetURL(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "/storage/") || strings.HasPrefix(trimmed, "storage/") {
		cleanPath := strings.TrimPrefix(trimmed, "/")
		if s.contentAssetBase == "" {
			return "/" + cleanPath
		}
		return s.contentAssetBase + "/" + cleanPath
	}
	if s.contentAssetBase == "" {
		return "/storage/" + strings.TrimPrefix(trimmed, "/")
	}
	return s.contentAssetBase + "/storage/" + strings.TrimPrefix(trimmed, "/")
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 10
	}
	if pageSize > 50 {
		return 50
	}
	return pageSize
}

func firstNonEmptyStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func IsContentNewsNotFound(err error) bool {
	return errors.Is(err, repopg.ErrContentNewsNotFound)
}
