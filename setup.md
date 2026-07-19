# FH88U Production Setup Guide

Tài liệu này viết lại từ đầu quy trình dựng production cho bộ source trong repo này trên **1 VPS Ubuntu/Debian** theo hướng ổn định hơn bản cũ.

Mục tiêu:

- `vue` phục vụ app khách hàng
- `agency` phục vụ agency dashboard
- `gin-api` phục vụ API chính
- `gin-engine` chạy game engine / scheduler runtime
- `gate` phục vụ webhook / tích hợp ngoài
- `admin` chạy Laravel + Filament bằng `nginx + php-fpm`
- PostgreSQL + Redis chạy local trên cùng máy

Luu y quan trọng:

- Tài liệu này dành cho **single VPS**. Nó giúp hệ thống sạch và ổn định hơn, nhưng **không phải high availability**.
- Việc “gắn thêm domain” không giúp tăng chịu lỗi nếu tất cả domain vẫn trỏ về **cùng 1 VPS**.
- Không lưu secret thật vào repo hoặc vào tài liệu này.

---

## 1. Gia su va naming

Ví dụ domain theo zone Cloudflare hiện tại:

- App khách: `fh88u1.win`
- API: `api.fh88u1.win`
- Admin Laravel: `admin.fh88u1.win`
- Daily / agency dashboard: `daily.fh88u1.win`

Ví dụ đường dẫn source:

- source code: `/app`
- binary Go: `/app/bin`
- script deploy: `/app/scripts`

Ví dụ user chạy app:

- user system: `deploy`

Khuyến nghị version:

- Ubuntu 24.04 LTS hoặc Debian 12
- PHP `8.3`
- PostgreSQL `16`
- Redis `7`
- Node.js `22`
- pnpm `10.33.0`
- Go `1.26.1`

---

## 2. Bootstrap server

Đăng nhập `root`, cập nhật hệ thống:

```bash
apt update && apt upgrade -y
timedatectl set-timezone Asia/Ho_Chi_Minh
hostnamectl set-hostname fh88u-prod
```

Tạo user deploy:

```bash
adduser --disabled-password --gecos "" deploy
usermod -aG sudo deploy
mkdir -p /home/deploy/.ssh
chmod 700 /home/deploy/.ssh
cp /root/.ssh/authorized_keys /home/deploy/.ssh/authorized_keys
chown -R deploy:deploy /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys
```

Khuyến nghị bật swap để chống spike RAM:

```bash
fallocate -l 4G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
swapon --show
free -h
```

Firewall tối thiểu:

```bash
ufw allow OpenSSH
ufw allow 'Nginx Full'
ufw --force enable
ufw status
```

---

## 3. Cai package he thong

### 3.1 Package base

```bash
apt install -y \
  ca-certificates curl gnupg lsb-release unzip zip jq git rsync \
  build-essential software-properties-common acl \
  nginx redis-server postgresql postgresql-contrib \
  imagemagick
```

### 3.2 PHP 8.3 + extensions

```bash
add-apt-repository ppa:ondrej/php -y
apt update
apt install -y \
  php8.3 php8.3-cli php8.3-fpm php8.3-common \
  php8.3-bcmath php8.3-curl php8.3-gd php8.3-intl php8.3-mbstring \
  php8.3-pgsql php8.3-redis php8.3-xml php8.3-zip php8.3-opcache
```

Kiểm tra:

```bash
php -v
php -m | egrep 'bcmath|curl|gd|intl|mbstring|pgsql|redis|xml|zip'
systemctl enable --now php8.3-fpm
systemctl status php8.3-fpm --no-pager
```

### 3.3 Composer

```bash
cd /tmp
php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');"
php composer-setup.php --install-dir=/usr/local/bin --filename=composer
rm -f composer-setup.php
composer --version
```

### 3.4 Node.js 22 + pnpm

```bash
curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
apt install -y nodejs
npm install -g corepack
corepack enable
corepack prepare pnpm@10.33.0 --activate
node -v
npm -v
pnpm -v
```

### 3.5 Go 1.26.1

```bash
cd /tmp
curl -LO https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
cat >/etc/profile.d/go.sh <<'EOF'
export PATH=/usr/local/go/bin:$PATH
EOF
chmod 644 /etc/profile.d/go.sh
source /etc/profile.d/go.sh
go version
```

---

## 4. Tao structure thu muc

```bash
mkdir -p /app/bin /app/scripts /app/storage/backups
chown -R deploy:deploy /app
```

Clone source:

```bash
sudo -u deploy git clone <YOUR_GIT_REPO_URL> /app
cd /app
sudo -u deploy git branch
```

---

## 5. PostgreSQL

Bật dịch vụ:

```bash
systemctl enable --now postgresql
systemctl status postgresql --no-pager
pg_lsclusters
```

Tạo database và user:

```bash
sudo -u postgres psql <<'EOF'
CREATE USER fh88u WITH PASSWORD 'CHANGE_ME_STRONG_DB_PASSWORD';
CREATE DATABASE fh88u OWNER fh88u;
GRANT ALL PRIVILEGES ON DATABASE fh88u TO fh88u;
\q
EOF
```

Test kết nối:

```bash
PGPASSWORD='CHANGE_ME_STRONG_DB_PASSWORD' psql \
  -h 127.0.0.1 -U fh88u -d fh88u -c 'select now();'
```

---

## 6. Redis

Bật dịch vụ:

```bash
systemctl enable --now redis-server
systemctl status redis-server --no-pager
redis-cli ping
```

Gợi ý chia Redis DB:

- DB `0`: queue/session/cache mặc định cho Laravel
- DB `1`: cache phụ nếu cần
- DB `2`: shared snapshot cho `gin` / `gate` / `admin`

Điểm này rất quan trọng:

- `admin` đang dùng `REDIS_SHARED_DB=2`
- `gate` nên dùng `SHARED_REDIS_DB=2`
- `gin` nên dùng `REDIS_DB=2` để đọc đúng shared keys

---

## 7. Tao file env

### 7.1 `gin/.env`

```env
APP_NAME=gin-core
HTTP_ADDR=:8081
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s
HTTP_SHUTDOWN_TIMEOUT=10s

CONTENT_ASSET_BASE_URL=https://admin.fh88u1.win
POPUP_VIDEO_FILE_PATH=/app/vue/public/bg.mp4

DATABASE_URL=postgresql://fh88u:CHANGE_ME_STRONG_DB_PASSWORD@127.0.0.1:5432/fh88u

AUTH_TOKEN_SECRET=CHANGE_ME_STRONG_AUTH_SECRET
AUTH_TOKEN_TTL=2h
AUTH_REFRESH_TOKEN_TTL=720h
PUBLIC_REGISTER_URL=https://fh88u1.win/register

REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=2

GATE_BASE_URL=http://127.0.0.1:8082
GIN_INTERNAL_TOKEN=CHANGE_ME_INTERNAL_TOKEN
PAYMENT_RECEIVING_ACCOUNTS_REDIS_KEY=shared:payment:receiving-accounts:v1

AUTH_FORGOT_OTP_TTL=2m
AUTH_FORGOT_OTP_COOLDOWN=2m
AUTH_FORGOT_OTP_MAX_ATTEMPTS=5
AUTH_FORGOT_WINDOW=2m
AUTH_FORGOT_LIMIT_IP=20
AUTH_FORGOT_LIMIT_TARGET=10
AUTH_LOGIN_FAIL_WINDOW=2m
AUTH_LOGIN_FAIL_LIMIT_IP=30
AUTH_LOGIN_FAIL_LIMIT_ACCOUNT=10
AUTH_LOGIN_LOCK_DURATION=2m
AUTH_REGISTER_WINDOW=2m
AUTH_REGISTER_LIMIT_IP=20
AUTH_REGISTER_LIMIT_EMAIL=10
AUTH_REGISTER_LIMIT_PHONE=10

TCG_ENABLED=true
TCG_DEFAULT_LANGUAGE=VI
TCG_DEFAULT_GAME_MODE=1
TCG_DEFAULT_IP_ADDRESS=127.0.0.1
TCG_DEFAULT_WEB_PLATFORM=html5-desktop
TCG_DEFAULT_MOBILE_PLATFORM=html5
TCG_DEFAULT_LOTTERY_LOBBY_GAME_CODE=Lobby
TCG_LAUNCH_RETURN_URL=https://fh88u1.win/play
TCG_DEFAULT_CURRENCY=VND
TCG_GAME_LIST_REDIS_KEY=shared:tcg:game-list:v1
TCG_PREVIEW_ZERO_BALANCE=false
```

### 7.2 `gate/.env`

```env
APP_NAME=gate
HTTP_ADDR=:8082
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s
HTTP_SHUTDOWN_TIMEOUT=10s

GIN_INTERNAL_BASE_URL=http://127.0.0.1:8081
GIN_INTERNAL_TOKEN=CHANGE_ME_INTERNAL_TOKEN
GATE_INTERNAL_TOKEN=CHANGE_ME_INTERNAL_TOKEN

SHARED_REDIS_ADDR=127.0.0.1:6379
SHARED_REDIS_PASSWORD=
SHARED_REDIS_DB=2
EXCHANGE_RATE_REDIS_KEY=shared:exchange-rate:usdt-vnd
GATE_PUBLIC_BASE_URL=https://api.fh88u1.win

NOWPAYMENTS_BASE_URL=https://api.nowpayments.io
NOWPAYMENTS_API_KEY=CHANGE_ME_NOWPAYMENTS_API_KEY
NOWPAYMENTS_IPN_SECRET=CHANGE_ME_NOWPAYMENTS_IPN_SECRET
NOWPAYMENTS_PRICE_CURRENCY=usd
NOWPAYMENTS_PAY_CURRENCY=USDTTRC20

TCG_ENABLED=true
TCG_BASE_URL=CHANGE_ME_TCG_BASE_URL
TCG_HTTP_TIMEOUT=30s
TCG_MERCHANT_CODE=CHANGE_ME
TCG_MERCHANT_DES_KEY=CHANGE_ME
TCG_MERCHANT_SIGN_KEY=CHANGE_ME
TCG_REPORT_FTP_HOST=CHANGE_ME
TCG_REPORT_FTP_PORT=21
TCG_REPORT_FTP_USERNAME=CHANGE_ME
TCG_REPORT_FTP_PASSWORD=CHANGE_ME
TCG_REPORT_FTP_BASE_DIR=CHANGE_ME
TCG_GAME_LIST_SYNC_ENABLED=true
TCG_GAME_LIST_SYNC_INTERVAL=30s
TCG_GAME_LIST_REDIS_KEY=shared:tcg:game-list:v1
TCG_GAME_LIST_PLATFORM=all
TCG_GAME_LIST_CLIENT_TYPE=all
TCG_GAME_LIST_LANGUAGE=VI
TCG_GAME_LIST_PAGE=0
TCG_GAME_LIST_PAGE_SIZE=0
TCG_GAME_LIST_PRODUCTS_RNG=275,16,141,55,140,98,243,212
TCG_GAME_LIST_PRODUCTS_FISH=79,16,141,55,140,43,138,243,212
TCG_GAME_LIST_PRODUCTS_LIVE=4,112,3,79,27,177,43,39,93,258,118,272
TCG_GAME_LIST_PRODUCTS_PVP=140,43,212
TCG_GAME_LIST_PRODUCTS_SPORTS=104,65,68,131,174,202,132
TCG_GAME_LIST_PRODUCTS_ELOTT=2,384,420,64,76
```

### 7.3 `admin/.env`

Lưu ý quan trọng: dùng `DB_CONNECTION=pgsql`, không dùng `postgres`.

```env
APP_NAME=fh88u-admin
APP_ENV=production
APP_KEY=
APP_DEBUG=false
APP_URL=https://admin.fh88u1.win

APP_LOCALE=vi
APP_FALLBACK_LOCALE=vi
APP_FAKER_LOCALE=vi_VN

HOST_MOBILE=https://fh88u1.win
VUE_ADMIN_CONTROL_URL=https://fh88u1.win/auth/sso

APP_MAINTENANCE_DRIVER=file
BCRYPT_ROUNDS=12

LOG_CHANNEL=stack
LOG_STACK=single
LOG_LEVEL=info
LOG_VIEWER_ENABLED=false
LOG_VIEWER_REQUIRE_AUTH_IN_PRODUCTION=true

DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=fh88u
DB_USERNAME=fh88u
DB_PASSWORD=CHANGE_ME_STRONG_DB_PASSWORD

SESSION_DRIVER=redis
SESSION_LIFETIME=120
SESSION_ENCRYPT=false
SESSION_PATH=/
SESSION_DOMAIN=.fh88u1.win

BROADCAST_CONNECTION=log
FILESYSTEM_DISK=public
QUEUE_CONNECTION=redis
CACHE_STORE=redis

REDIS_CLIENT=phpredis
REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379
REDIS_DB=0
REDIS_CACHE_DB=1
REDIS_SHARED_DB=2

USDT_VND_RATE_URL=https://api.coingecko.com/api/v3/simple/price?ids=tether&vs_currencies=vnd
USDT_VND_RATE_SOURCE_NAME=coingecko
USDT_VND_RATE_TIMEOUT=10
USDT_VND_RATE_RETRY=2
USDT_VND_RATE_CACHE_TTL_SECONDS=300
USDT_VND_RATE_CACHE_STORE=redis
USDT_VND_RATE_CACHE_KEY=admin:exchange-rate:usdt-vnd:snapshot
USDT_VND_RATE_REDIS_CONNECTION=shared
USDT_VND_RATE_REDIS_KEY=shared:exchange-rate:usdt-vnd

VIETQR_BANKS_URL=https://api.vietqr.io/v2/banks
VIETQR_BANKS_SOURCE_NAME=vietqr
VIETQR_BANKS_TIMEOUT=10
VIETQR_BANKS_RETRY=2
VIETQR_BANKS_CACHE_TTL_SECONDS=86400
VIETQR_BANKS_CACHE_STORE=redis
VIETQR_BANKS_CACHE_KEY=admin:vietqr:banks:snapshot
VIETQR_BANKS_REDIS_CONNECTION=shared
VIETQR_BANKS_REDIS_KEY=shared:vietqr:banks

PAYMENT_RECEIVING_ACCOUNTS_SOURCE_NAME=payment_receiving_accounts
PAYMENT_RECEIVING_ACCOUNTS_CACHE_STORE=redis
PAYMENT_RECEIVING_ACCOUNTS_CACHE_KEY=admin:payment:receiving-accounts:snapshot
PAYMENT_RECEIVING_ACCOUNTS_CACHE_TTL_SECONDS=300
PAYMENT_RECEIVING_ACCOUNTS_REDIS_CONNECTION=shared
PAYMENT_RECEIVING_ACCOUNTS_REDIS_KEY=shared:payment:receiving-accounts:v1

TCG_GATE_BASE_URL=http://127.0.0.1:8082
TCG_GATE_INTERNAL_TOKEN=CHANGE_ME_INTERNAL_TOKEN
TCG_GATE_TIMEOUT=15
TCG_DEFAULT_CURRENCY=VND

MAIL_MAILER=log
MAIL_FROM_ADDRESS=hello@fh88u1.win
MAIL_FROM_NAME="${APP_NAME}"
```

### 7.4 `vue/.env`

```env
VITE_API_BASE_URL=https://api.fh88u1.win
VITE_ALLOWED_HOSTS=https://fh88u1.win
DEV=false
```

### 7.5 `agency/.env`

```env
VITE_API_BASE_URL=https://api.fh88u1.win
VITE_ALLOWED_HOSTS=https://daily.fh88u1.win
```

---

## 8. Cai dependency va build lan dau

Chạy toàn bộ bằng user `deploy`:

```bash
sudo -u deploy bash <<'EOF'
set -e
source /etc/profile

cd /app/gin
go mod download
go build -o /app/bin/gin-api ./cmd/api
go build -o /app/bin/gin-engine ./cmd/engine

cd /app/gate
go mod download
go build -o /app/bin/gate ./cmd/...

cd /app/vue
pnpm install --frozen-lockfile
pnpm run build

cd /app/agency
pnpm install --frozen-lockfile
pnpm run build

cd /app/admin
composer install --no-dev --optimize-autoloader
npm install
npm run build

php artisan key:generate --force
php artisan storage:link || true
php artisan optimize:clear
php artisan migrate --force
php artisan config:cache
php artisan route:cache
php artisan view:cache
php artisan filament:optimize
EOF
```

Fix quyền cho Laravel:

```bash
chown -R deploy:www-data /app/admin/storage /app/admin/bootstrap/cache
chmod -R ug+rwx /app/admin/storage /app/admin/bootstrap/cache
setfacl -R -m u:www-data:rwx -m u:deploy:rwx /app/admin/storage /app/admin/bootstrap/cache
setfacl -dR -m u:www-data:rwx -m u:deploy:rwx /app/admin/storage /app/admin/bootstrap/cache
```

---

## 9. Systemd services

### 9.1 `gin-api.service`

```bash
cat >/etc/systemd/system/gin-api.service <<'EOF'
[Unit]
Description=FH88U Gin API
After=network.target postgresql.service redis-server.service

[Service]
Type=simple
User=deploy
WorkingDirectory=/app/gin
ExecStart=/app/bin/gin-api
Restart=always
RestartSec=5
EnvironmentFile=/app/gin/.env
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

### 9.2 `gin-engine.service`

```bash
cat >/etc/systemd/system/gin-engine.service <<'EOF'
[Unit]
Description=FH88U Game Engine
After=network.target postgresql.service redis-server.service

[Service]
Type=simple
User=deploy
WorkingDirectory=/app/gin
ExecStart=/app/bin/gin-engine
Restart=always
RestartSec=5
EnvironmentFile=/app/gin/.env
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

### 9.3 `gate.service`

```bash
cat >/etc/systemd/system/gate.service <<'EOF'
[Unit]
Description=FH88U Gate Service
After=network.target redis-server.service

[Service]
Type=simple
User=deploy
WorkingDirectory=/app/gate
ExecStart=/app/bin/gate
Restart=always
RestartSec=5
EnvironmentFile=/app/gate/.env
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

### 9.4 Laravel queue worker

```bash
cat >/etc/systemd/system/fh88u-queue.service <<'EOF'
[Unit]
Description=FH88U Laravel Queue Worker
After=network.target redis-server.service php8.3-fpm.service

[Service]
Type=simple
User=deploy
WorkingDirectory=/app/admin
ExecStart=/usr/bin/php artisan queue:work redis --sleep=1 --tries=3 --timeout=120 --queue=default
Restart=always
RestartSec=5
EnvironmentFile=/app/admin/.env
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

### 9.5 Laravel scheduler

Quan trọng: repo đang có `everyFiveSeconds()` trong [admin/routes/console.php](/home/truonghocdot/study/practice/admin/routes/console.php), nên production phải chạy `schedule:work`, không dùng cron `schedule:run`.

```bash
cat >/etc/systemd/system/fh88u-scheduler.service <<'EOF'
[Unit]
Description=FH88U Laravel Scheduler
After=network.target redis-server.service php8.3-fpm.service

[Service]
Type=simple
User=deploy
WorkingDirectory=/app/admin
ExecStart=/usr/bin/php artisan schedule:work
Restart=always
RestartSec=5
EnvironmentFile=/app/admin/.env
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

Bật toàn bộ service:

```bash
systemctl daemon-reload
systemctl enable --now gin-api gin-engine gate fh88u-queue fh88u-scheduler php8.3-fpm redis-server postgresql nginx
systemctl status gin-api gin-engine gate fh88u-queue fh88u-scheduler php8.3-fpm nginx --no-pager
```

---

## 10. Nginx

### 10.1 App khách `fh88u1.win`

```bash
cat >/etc/nginx/sites-available/fh88u-app.conf <<'EOF'
server {
    listen 80;
    server_name fh88u1.win www.fh88u1.win;

    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name fh88u1.win www.fh88u1.win;

    ssl_certificate /etc/ssl/cloudflare/fh88u1.win-origin.crt;
    ssl_certificate_key /etc/ssl/cloudflare/fh88u1.win-origin.key;

    root /app/vue/dist;
    index index.html;

    client_max_body_size 20m;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
EOF
```

### 10.2 Daily / agency `daily.fh88u1.win`

```bash
cat >/etc/nginx/sites-available/fh88u-agency.conf <<'EOF'
server {
    listen 80;
    server_name daily.fh88u1.win;

    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name daily.fh88u1.win;

    ssl_certificate /etc/ssl/cloudflare/fh88u1.win-origin.crt;
    ssl_certificate_key /etc/ssl/cloudflare/fh88u1.win-origin.key;

    root /app/agency/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
EOF
```

### 10.3 API `api.fh88u1.win`

```bash
cat >/etc/nginx/sites-available/fh88u-api.conf <<'EOF'
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

server {
    listen 80;
    server_name api.fh88u1.win;

    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.fh88u1.win;

    ssl_certificate /etc/ssl/cloudflare/fh88u1.win-origin.crt;
    ssl_certificate_key /etc/ssl/cloudflare/fh88u1.win-origin.key;

    client_max_body_size 20m;

    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_read_timeout 300;
    }
}
EOF
```

### 10.4 Admin Laravel `admin.fh88u1.win`

```bash
cat >/etc/nginx/sites-available/fh88u-admin.conf <<'EOF'
server {
    listen 80;
    server_name admin.fh88u1.win;

    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name admin.fh88u1.win;

    ssl_certificate /etc/ssl/cloudflare/fh88u1.win-origin.crt;
    ssl_certificate_key /etc/ssl/cloudflare/fh88u1.win-origin.key;

    root /app/admin/public;
    index index.php index.html;

    client_max_body_size 50m;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        fastcgi_param DOCUMENT_ROOT $realpath_root;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }
}
EOF
```

Bật site:

```bash
ln -sf /etc/nginx/sites-available/fh88u-app.conf /etc/nginx/sites-enabled/fh88u-app.conf
ln -sf /etc/nginx/sites-available/fh88u-agency.conf /etc/nginx/sites-enabled/fh88u-agency.conf
ln -sf /etc/nginx/sites-available/fh88u-api.conf /etc/nginx/sites-enabled/fh88u-api.conf
ln -sf /etc/nginx/sites-available/fh88u-admin.conf /etc/nginx/sites-enabled/fh88u-admin.conf
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx
```

---

## 11. Cloudflare DNS + SSL

Bạn đang dùng Cloudflare proxied cho:

- `fh88u1.win`
- `admin.fh88u1.win`
- `api.fh88u1.win`
- `daily.fh88u1.win`

Khuyến nghị:

- tất cả record để **Proxied** (orange cloud)
- trong Cloudflare `SSL/TLS` đặt mode là **Full (strict)**
- trên origin dùng **Cloudflare Origin CA**

Tài liệu chính thức:

- Origin CA: https://developers.cloudflare.com/ssl/origin-configuration/origin-ca/
- Full (strict): https://developers.cloudflare.com/ssl/origin-configuration/ssl-modes/full-strict/

### 11.1 Tạo DNS trên Cloudflare

Tạo record A:

- `fh88u1.win` -> `109.123.236.99`
- `admin.fh88u1.win` -> `109.123.236.99`
- `api.fh88u1.win` -> `109.123.236.99`
- `daily.fh88u1.win` -> `109.123.236.99`

```bash
# kiểm tra DNS về đúng IP origin
dig +short fh88u1.win
dig +short admin.fh88u1.win
dig +short api.fh88u1.win
dig +short daily.fh88u1.win
```

### 11.2 Tạo Origin Certificate

Trong Cloudflare Dashboard:

- vào `SSL/TLS`
- chọn `Origin Server`
- `Create Certificate`
- SAN:
  - `fh88u1.win`
  - `*.fh88u1.win`

Lưu cert và key lên server:

```bash
mkdir -p /etc/ssl/cloudflare
chmod 700 /etc/ssl/cloudflare

cat >/etc/ssl/cloudflare/fh88u1.win-origin.crt <<'EOF'
PASTE_CLOUDFLARE_ORIGIN_CERT_HERE
EOF

cat >/etc/ssl/cloudflare/fh88u1.win-origin.key <<'EOF'
PASTE_CLOUDFLARE_ORIGIN_PRIVATE_KEY_HERE
EOF

chmod 644 /etc/ssl/cloudflare/fh88u1.win-origin.crt
chmod 600 /etc/ssl/cloudflare/fh88u1.win-origin.key

nginx -t
systemctl reload nginx
```

### 11.3 Bật SSL mode Full (strict)

Trong Cloudflare:

- `SSL/TLS` -> `Overview`
- Chọn `Full (strict)`

### 11.4 Khuyến nghị bảo vệ origin

Sau khi mọi thứ ổn định, nên giới hạn origin chỉ nhận traffic từ IP Cloudflare hoặc bật Authenticated Origin Pulls.

Tài liệu:

- https://developers.cloudflare.com/fundamentals/security/protect-your-origin-server/
- https://developers.cloudflare.com/ssl/origin-configuration/authenticated-origin-pull/

---

## 12. Deploy script production

Tạo file `/app/scripts/deploy.sh`:

```bash
cat >/app/scripts/deploy.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

source /etc/profile

echo "==> Pull latest code"
cd /app
sudo -u deploy git pull --ff-only origin main

echo "==> Build gin"
cd /app/gin
sudo -u deploy go mod download
sudo -u deploy go build -o /app/bin/gin-api ./cmd/api
sudo -u deploy go build -o /app/bin/gin-engine ./cmd/engine

echo "==> Build gate"
cd /app/gate
sudo -u deploy go mod download
sudo -u deploy go build -o /app/bin/gate ./cmd/...

echo "==> Build vue"
cd /app/vue
sudo -u deploy pnpm install --frozen-lockfile
sudo -u deploy pnpm run build

echo "==> Build agency"
cd /app/agency
sudo -u deploy pnpm install --frozen-lockfile
sudo -u deploy pnpm run build

echo "==> Build admin"
cd /app/admin
sudo -u deploy composer install --no-dev --optimize-autoloader
sudo -u deploy npm install
sudo -u deploy npm run build
sudo -u deploy php artisan optimize:clear
sudo -u deploy php artisan migrate --force
sudo -u deploy php artisan config:cache
sudo -u deploy php artisan route:cache
sudo -u deploy php artisan view:cache
sudo -u deploy php artisan filament:optimize

echo "==> Fix permissions"
chown -R deploy:www-data /app/admin/storage /app/admin/bootstrap/cache
chmod -R ug+rwx /app/admin/storage /app/admin/bootstrap/cache

echo "==> Restart services"
systemctl restart gin-api gin-engine gate php8.3-fpm fh88u-queue fh88u-scheduler
systemctl reload nginx

echo "==> Health check"
systemctl is-active gin-api gin-engine gate php8.3-fpm fh88u-queue fh88u-scheduler nginx
curl -fsS http://127.0.0.1:8081/healthz
EOF
chmod +x /app/scripts/deploy.sh
```

Deploy:

```bash
/app/scripts/deploy.sh
```

---

## 13. Kiem tra sau deploy

```bash
systemctl status gin-api gin-engine gate php8.3-fpm fh88u-queue fh88u-scheduler nginx --no-pager
curl -I https://fh88u1.win
curl -I https://daily.fh88u1.win
curl -I https://admin.fh88u1.win
curl -s https://api.fh88u1.win/healthz
```

Kiểm tra cổng:

```bash
ss -ltnp | grep -E ':80|:443|:8081|:8082|:5432|:6379'
```

---

## 14. Log va debug nhanh

```bash
journalctl -u gin-api -f
journalctl -u gin-engine -f
journalctl -u gate -f
journalctl -u php8.3-fpm -f
journalctl -u fh88u-queue -f
journalctl -u fh88u-scheduler -f
tail -f /app/admin/storage/logs/laravel.log
tail -f /var/log/nginx/error.log
tail -f /var/log/postgresql/postgresql-16-main.log
```

Khi nghi hệ thống chết toàn tập:

```bash
date
systemctl status postgresql redis-server gin-api gin-engine gate php8.3-fpm fh88u-queue fh88u-scheduler nginx --no-pager
pg_lsclusters
dmesg -T | grep -iE 'oom|killed process|out of memory|kill' || true
df -h
free -h
ss -ltnp | grep -E ':80|:443|:8081|:8082|:5432|:6379'
```

---

## 15. Backup toi thieu

Tạo thư mục backup:

```bash
mkdir -p /app/storage/backups/postgres
chown -R deploy:deploy /app/storage/backups
```

Script backup DB:

```bash
cat >/app/scripts/backup-postgres.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

STAMP="$(date +%F-%H%M%S)"
export PGPASSWORD='CHANGE_ME_STRONG_DB_PASSWORD'

pg_dump -h 127.0.0.1 -U fh88u -d fh88u \
  | gzip > "/app/storage/backups/postgres/fh88u-${STAMP}.sql.gz"

find /app/storage/backups/postgres -type f -name '*.sql.gz' -mtime +7 -delete
EOF
chmod +x /app/scripts/backup-postgres.sh
```

Cron chạy mỗi đêm:

```bash
cat >/etc/cron.d/fh88u-backup <<'EOF'
15 2 * * * root /app/scripts/backup-postgres.sh >> /var/log/fh88u-backup.log 2>&1
EOF
```

---

## 16. Khong lam trong production

Không nên:

- chạy `php artisan serve` làm web server production
- trộn secret thật vào repo hoặc `setup.md`
- cho `gin`, `gate`, `admin`, `postgres`, `redis` dùng chung 1 credential yếu
- dùng thêm domain như một cách “tăng chịu lỗi” nếu vẫn trỏ cùng một VPS

---

## 17. Ghi chu van hanh

Nếu cần reset game periods:

```bash
cd /app/admin
sudo -u deploy php artisan bet:reset-game-periods --force
```

Chỉ chạy khi bạn thực sự muốn reset nghiệp vụ, không đưa vào deploy mặc định.

---

## 18. Huong nang cap tiep theo

Sau khi chạy ổn trên single VPS, thứ tự nâng cấp nên là:

1. Tách PostgreSQL ra managed DB hoặc máy riêng
2. Tách Redis ra managed Redis hoặc máy riêng
3. Thêm monitoring + alert Telegram/Uptime
4. Có thêm 1 app node dự phòng
5. Đặt load balancer / health-check failover

Khi đó việc “gắn thêm domain” mới có ý nghĩa vận hành thực sự.
