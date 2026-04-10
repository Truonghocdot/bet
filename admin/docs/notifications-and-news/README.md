# Notifications (In-App) + News (Tin Tuc) - Spec

Tài liệu này là **nghiệp vụ + hợp đồng kỹ thuật** cho 2 module:

- **Notifications (In-App, không realtime)**: Admin tạo thông báo trong ERP, ERP gọi `gate`, `gate` ghi DB. End-user app (qua `gin`) chỉ đọc DB và đánh dấu đã đọc.
- **News (Tin tức/bài viết)**: Admin CRUD bài viết trực tiếp trong Laravel/Filament (Laravel ghi DB). End-user app (qua `gin`) chỉ đọc DB để hiển thị.

Mục tiêu: giảm phụ thuộc realtime, giữ DB là source-of-truth, phân vai rõ `ERP / gate / gin`.

## 1) Phân vai

### 1.1 ERP (Laravel/Filament)

- UI quản trị tạo thông báo (notification).
- UI quản trị đăng bài viết tin tức (news post).
- Khi admin "Gửi/Publish" notification: gọi API nội bộ của `gate`.

### 1.2 Gate

- **Writer** của notifications: validate + transaction ghi DB `notifications` và `notification_targets`.
- Không phục vụ end-user read notifications.
- Có auth nội bộ bằng `GATE_INTERNAL_TOKEN`.

### 1.3 Gin

- **Reader** cho notifications/news:
  - list notifications cho user (kèm `is_read`)
  - mark read (idempotent)
  - list news, get news detail

## 2) Data Model (DB)

### 2.1 Notifications

Thiết kế 3 bảng để hỗ trợ:

- broadcast toàn hệ thống (không nổ dữ liệu)
- nhắm tới danh sách user (targeted)
- theo dõi đã đọc

#### `notifications`

- `id`: bigint
- `title`: varchar(200)
- `body`: text
- `status`: tinyint
  - `DRAFT = 1`
  - `PUBLISHED = 2`
  - `ARCHIVED = 3`
- `audience`: tinyint
  - `ALL = 1`
  - `USERS = 2`
- `publish_at`: timestamp nullable
  - `null` nghĩa là publish ngay khi `status = PUBLISHED`
- `expires_at`: timestamp nullable
- `created_by`: bigint nullable (admin user id)
- `created_at`, `updated_at`

Index:

- `index(status, publish_at)`
- `index(audience, status)`

Rule:

- `PUBLISHED` + `publish_at <= now` (hoặc `publish_at is null`) mới hiển thị cho end-user.
- `expires_at` nếu có thì `expires_at > now` mới hiển thị.

#### `notification_targets` (chỉ dùng khi `audience=USERS`)

- `id`: bigint
- `notification_id`: bigint FK -> `notifications.id`
- `user_id`: bigint FK -> `users.id`
- `created_at`

Constraint:

- unique(`notification_id`, `user_id`)

Index:

- `index(user_id, notification_id)`

#### `notification_reads`

- `id`: bigint
- `notification_id`: bigint FK -> `notifications.id`
- `user_id`: bigint FK -> `users.id`
- `read_at`: timestamp

Constraint:

- unique(`notification_id`, `user_id`) (để mark read idempotent)

Index:

- `index(user_id, read_at)`

### 2.2 News posts

#### `news_posts`

- `id`: bigint
- `slug`: varchar(200) unique
- `title`: varchar(255)
- `excerpt`: text nullable
- `content`: longtext
- `cover_image`: varchar(255) nullable
- `status`: tinyint
  - `DRAFT = 1`
  - `PUBLISHED = 2`
  - `ARCHIVED = 3`
- `published_at`: timestamp nullable
  - nếu publish mà trống thì set `now()`
- `created_by`: bigint nullable
- `updated_by`: bigint nullable
- `created_at`, `updated_at`

Index:

- `index(status, published_at)`
- `index(created_by)`

## 3) Query rule cho Gin

### 3.1 List notifications cho user

Điều kiện chung:

- `status = PUBLISHED`
- `publish_at is null OR publish_at <= now`
- `expires_at is null OR expires_at > now`

Broadcast:

- `audience = ALL`

Targeted:

- `audience = USERS`
- join `notification_targets` theo `user_id = current_user`

Read state:

- left join `notification_reads` theo `notification_id + user_id`
- trả `is_read = (notification_reads.read_at != null)`

### 3.2 Mark read idempotent

- upsert theo unique(`notification_id`, `user_id`)
- nếu đã tồn tại thì không đổi `read_at` (hoặc cập nhật lại `read_at`, tùy policy; khuyến nghị: giữ lần đầu).

### 3.3 List news

Public:

- `status = PUBLISHED`
- `published_at is not null AND published_at <= now`

## 4) API Contracts

### 4.1 Gate API (writer)

#### `POST /v1/inapp-notifications`

Headers:

- `Authorization: Bearer <GATE_INTERNAL_TOKEN>`
- `Content-Type: application/json`

Request JSON:

- `title` (required)
- `body` (required)
- `audience` (`ALL` | `USERS`) (required)
- `target_user_ids` (optional; required khi `USERS`)
- `publish_at` (optional)
- `expires_at` (optional)
- `created_by` (optional)

Behavior:

- validate input
- transaction:
  - insert `notifications`
  - insert `notification_targets` nếu `USERS`
- response: `notification_id`, `status`, `audience`

Errors:

- `401` nếu token sai
- `422` nếu validate fail

#### `GET /v1/inapp-notifications` (optional)

Gate có thể không cần endpoint này nếu ERP đọc DB trực tiếp.

### 4.2 Gin API (reader)

#### `GET /v1/notifications` (auth required)

Response:

- list notifications:
  - `id`, `title`, `body`, `publish_at`, `expires_at`
  - `is_read`

#### `POST /v1/notifications/{id}/read` (auth required)

Behavior:

- upsert `notification_reads`
- idempotent

Response:

- `{ message: "Đã đánh dấu đã đọc" }` (text cụ thể do gin quyết định)

#### `GET /v1/news` (public)

Response:

- list news:
  - `slug`, `title`, `excerpt`, `cover_image`, `published_at`

#### `GET /v1/news/{slug}` (public)

Response:

- `slug`, `title`, `content`, `cover_image`, `published_at`

## 5) ERP UI (Filament)

### 5.1 Notification UI

- Form:
  - `title`, `body`
  - `audience (ALL/USERS)`
  - `target users` (multi-select, chỉ hiện khi USERS)
  - `publish_at`, `expires_at`
- Action:
  - `Lưu nháp` (status=DRAFT, chỉ lưu local nếu ERP quản lý DB notification; nếu writer là gate thì ERP có thể chỉ có trạng thái local trước khi gọi gate)
  - `Gửi/Publish` (gọi `gate`)
- List:
  - filter `status`, `audience`
  - search `title`

### 5.2 News UI

- CRUD trực tiếp DB:
  - `title`, `slug` (auto), `excerpt`, `content`, `cover_image`, `status`, `published_at`
- Publish:
  - set `status=PUBLISHED`
  - set `published_at` nếu trống

## 6) Security

- `gate` phải verify `GATE_INTERNAL_TOKEN` cho endpoint writer.
- Token nên đặt ở env:
  - `GATE_INTERNAL_TOKEN` trên `gate`
  - `GATE_INTERNAL_TOKEN` trên `admin` (client) hoặc secret manager
- Không expose endpoint writer ra public internet nếu có thể (đặt sau private network / firewall).

## 7) Test Scenarios (Acceptance)

Gate:

- tạo broadcast -> DB có `notifications`, không có `notification_targets`
- tạo targeted -> DB có đúng số record `notification_targets`
- token sai -> `401`

Gin notifications:

- user A thấy broadcast + targeted của mình, không thấy targeted của user B
- mark read 2 lần không tạo record trùng (idempotent)

News:

- chỉ trả `PUBLISHED` và `published_at <= now`
- `slug` unique

## 8) Defaults (locked)

- Notifications: `broadcast + chọn user`, **chỉ in-app**, writer là `gate`.
- News: writer là **Laravel/Filament**, `gin` chỉ đọc.
- Không realtime: end-user polling theo màn.

