package content

import "time"

type BannerItem struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
	LinkURL  string `json:"link_url,omitempty"`
}

type NewsItem struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt,omitempty"`
	Content     string     `json:"content,omitempty"`
	CoverImage  string     `json:"cover_image_url,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type HomeResponse struct {
	Message    string       `json:"message"`
	Banners    []BannerItem `json:"banners"`
	Highlights []NewsItem   `json:"highlights"`
}

type PromotionListResponse struct {
	Message    string     `json:"message"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	Total      int        `json:"total"`
	TotalPages int        `json:"total_pages"`
	Items      []NewsItem `json:"items"`
}

type NewsListResponse struct {
	Message    string     `json:"message"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	Total      int        `json:"total"`
	TotalPages int        `json:"total_pages"`
	Items      []NewsItem `json:"items"`
}

type NewsDetailResponse struct {
	Message string     `json:"message"`
	Item    NewsItem   `json:"item"`
	Related []NewsItem `json:"related"`
}
