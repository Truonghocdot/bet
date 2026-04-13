package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type ContentRepository struct {
	db *sql.DB
}

var ErrContentNewsNotFound = errors.New("content.news.not_found")

type BannerRecord struct {
	ID        int64
	Title     string
	ImagePath string
	LinkURL   *string
	SortOrder int
}

type NewsRecord struct {
	ID          int64
	Title       string
	Slug        string
	Excerpt     *string
	Content     string
	CoverImage  *string
	PublishedAt *time.Time
	CreatedAt   time.Time
}

func NewContentRepository(db *sql.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

func (r *ContentRepository) ListActiveBanners(ctx context.Context, limit int, now time.Time) ([]BannerRecord, error) {
	if limit <= 0 {
		limit = 5
	}

	rows, err := r.db.QueryContext(ctx, `
		select
			id,
			title,
			image_path,
			link_url,
			sort_order
		from banners
		where deleted_at is null
		  and is_active = true
		  and (start_at is null or start_at <= $1)
		  and (end_at is null or end_at > $1)
		order by sort_order asc, id desc
		limit $2
	`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]BannerRecord, 0, limit)
	for rows.Next() {
		var item BannerRecord
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.ImagePath,
			&item.LinkURL,
			&item.SortOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *ContentRepository) ListPublishedNews(ctx context.Context, page, pageSize int, onlyPromotion bool) ([]NewsRecord, int, error) {
	page, pageSize = normalizePagination(page, pageSize)
	promoSQL := promotionWhereSQL()

	filter := "true"
	if onlyPromotion {
		filter = promoSQL
	} else {
		filter = "not (" + promoSQL + ")"
	}

	var total int
	countSQL := `
		select count(*)
		from news_articles
		where deleted_at is null
		  and is_published = true
		  and (published_at is null or published_at <= now())
		  and ` + filter
	if err := r.db.QueryRowContext(ctx, countSQL).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySQL := `
		select
			id,
			title,
			slug,
			excerpt,
			content,
			cover_image_path,
			published_at,
			created_at
		from news_articles
		where deleted_at is null
		  and is_published = true
		  and (published_at is null or published_at <= now())
		  and ` + filter + `
		order by coalesce(published_at, created_at) desc, id desc
		limit $1 offset $2
	`
	rows, err := r.db.QueryContext(ctx, querySQL, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]NewsRecord, 0, pageSize)
	for rows.Next() {
		var item NewsRecord
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Slug,
			&item.Excerpt,
			&item.Content,
			&item.CoverImage,
			&item.PublishedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *ContentRepository) ListLatestNews(ctx context.Context, limit int) ([]NewsRecord, error) {
	if limit <= 0 {
		limit = 6
	}

	rows, err := r.db.QueryContext(ctx, `
		select
			id,
			title,
			slug,
			excerpt,
			content,
			cover_image_path,
			published_at,
			created_at
		from news_articles
		where deleted_at is null
		  and is_published = true
		  and (published_at is null or published_at <= now())
		order by coalesce(published_at, created_at) desc, id desc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]NewsRecord, 0, limit)
	for rows.Next() {
		var item NewsRecord
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Slug,
			&item.Excerpt,
			&item.Content,
			&item.CoverImage,
			&item.PublishedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *ContentRepository) FindPublishedNewsBySlug(ctx context.Context, slug string) (NewsRecord, error) {
	var item NewsRecord
	err := r.db.QueryRowContext(ctx, `
		select
			id,
			title,
			slug,
			excerpt,
			content,
			cover_image_path,
			published_at,
			created_at
		from news_articles
		where deleted_at is null
		  and is_published = true
		  and (published_at is null or published_at <= now())
		  and slug = $1
		limit 1
	`, strings.TrimSpace(slug)).Scan(
		&item.ID,
		&item.Title,
		&item.Slug,
		&item.Excerpt,
		&item.Content,
		&item.CoverImage,
		&item.PublishedAt,
		&item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewsRecord{}, ErrContentNewsNotFound
		}
		return NewsRecord{}, err
	}

	return item, nil
}

func (r *ContentRepository) ListRelatedNews(ctx context.Context, slug string, limit int) ([]NewsRecord, error) {
	if limit <= 0 {
		limit = 3
	}

	rows, err := r.db.QueryContext(ctx, `
		select
			id,
			title,
			slug,
			excerpt,
			content,
			cover_image_path,
			published_at,
			created_at
		from news_articles
		where deleted_at is null
		  and is_published = true
		  and (published_at is null or published_at <= now())
		  and slug <> $1
		order by coalesce(published_at, created_at) desc, id desc
		limit $2
	`, strings.TrimSpace(slug), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]NewsRecord, 0, limit)
	for rows.Next() {
		var item NewsRecord
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Slug,
			&item.Excerpt,
			&item.Content,
			&item.CoverImage,
			&item.PublishedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func promotionWhereSQL() string {
	keywords := []string{
		"khuyến mãi",
		"ưu đãi",
		"thưởng",
		"hoàn trả",
		"hoàn tiền",
		"affiliate",
		"đại lý",
		"sự kiện",
		"bonus",
	}
	parts := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		escaped := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(keyword)), "'", "''")
		parts = append(parts, "lower(coalesce(title, '') || ' ' || coalesce(excerpt, '') || ' ' || coalesce(content, '')) like '%"+escaped+"%'")
	}
	return strings.Join(parts, " or ")
}
