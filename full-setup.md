# FH88U - Full setup Ubuntu 22.04 bằng root + Supervisor

Tài liệu này dành cho VPS Ubuntu `22.04` vừa khởi tạo. Toàn bộ source, build và process ứng dụng chạy bằng `root`; không tạo user `deploy`.

Các process ứng dụng chạy bằng Supervisor:

- `gin-api`
- `gin-engine`
- `gate`
- Laravel queue
- Laravel scheduler

Các daemon hạ tầng vẫn dùng user mặc định của package Ubuntu: Nginx/PHP-FPM, PostgreSQL, Redis và Supervisor.

## 0. Chuẩn bị DNS

Tạo 5 record `A` cùng trỏ về IP public của VPS:

| Domain | Thành phần |
| --- | --- |
| `fh88u.win` | App khách hàng |
| `admin.fh88u.win` | Laravel admin |
| `api.fh88u.win` | Gin API |
| `gate.fh88u.win` | Payment webhook |
| `daily.fh88u.win` | Agency dashboard |

Nếu dùng Cloudflare, nên chuyển 5 record sang `DNS only` trong lúc cấp SSL. Sau khi setup xong, bật lại `Proxied` và đặt `SSL/TLS` thành `Full (strict)`.

SSH vào VPS bằng `root`, rồi chạy lần lượt các phần dưới đây.

## 1. Cập nhật VPS

```bash
apt update
apt upgrade -y
timedatectl set-timezone Asia/Ho_Chi_Minh
hostnamectl set-hostname fh88u-prod
```

## 2. Tạo swap 2 GB

```bash
fallocate -l 2G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
```

## 3. Cài package hệ thống

```bash
apt install -y \
  acl build-essential ca-certificates certbot cron curl git gnupg imagemagick jq \
  lsb-release nginx openssl postgresql postgresql-contrib \
  python3-certbot-nginx redis-server rsync software-properties-common \
  supervisor ufw unzip zip
```

## 4. Cài PHP 8.3

```bash
add-apt-repository ppa:ondrej/php -y
apt update
apt install -y \
  php8.3 php8.3-bcmath php8.3-cli php8.3-common php8.3-curl php8.3-fpm \
  php8.3-gd php8.3-imagick php8.3-intl php8.3-mbstring php8.3-opcache \
  php8.3-pgsql php8.3-redis php8.3-xml php8.3-zip
```

Cấu hình PHP production:

```bash
mkdir -p /var/log/php
chown www-data:www-data /var/log/php

cat >/etc/php/8.3/fpm/conf.d/99-fh88u.ini <<'EOF'
memory_limit = 512M
max_execution_time = 300
max_input_time = 300
upload_max_filesize = 100M
post_max_size = 100M
expose_php = Off
display_errors = Off
log_errors = On
error_log = /var/log/php/error.log
opcache.enable = 1
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 16
opcache.max_accelerated_files = 20000
opcache.validate_timestamps = 0
EOF

systemctl enable --now php8.3-fpm
systemctl restart php8.3-fpm
```

## 5. Cài Composer

```bash
cd /tmp
EXPECTED_SIGNATURE="$(curl -fsSL https://composer.github.io/installer.sig)"
curl -fsSL https://getcomposer.org/installer -o composer-setup.php
ACTUAL_SIGNATURE="$(php -r "echo hash_file('sha384', 'composer-setup.php');")"
[ "$EXPECTED_SIGNATURE" = "$ACTUAL_SIGNATURE" ] || exit 1
php composer-setup.php --install-dir=/usr/local/bin --filename=composer
rm -f composer-setup.php
```

## 6. Cài Node.js 22 và pnpm

```bash
curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
apt install -y nodejs
npm install -g pnpm@10.33.0
```

## 7. Cài Go 1.26.1

VPS `amd64`:

```bash
cd /tmp
curl -fsSLO https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
echo '031f088e5d955bab8657ede27ad4e3bc5b7c1ba281f05f245bcc304f327c987a  go1.26.1.linux-amd64.tar.gz' | sha256sum -c -
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
rm -f go1.26.1.linux-amd64.tar.gz
echo 'export PATH=/usr/local/go/bin:$PATH' >/etc/profile.d/go.sh
export PATH=/usr/local/go/bin:$PATH
```

Nếu VPS là `arm64`, dùng block này thay cho block trên:

```bash
cd /tmp
curl -fsSLO https://go.dev/dl/go1.26.1.linux-arm64.tar.gz
echo 'a290581cfe4fe28ddd737dde3095f3dbeb7f2e4065cab4eae44dfc53b760c2f7  go1.26.1.linux-arm64.tar.gz' | sha256sum -c -
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.26.1.linux-arm64.tar.gz
rm -f go1.26.1.linux-arm64.tar.gz
echo 'export PATH=/usr/local/go/bin:$PATH' >/etc/profile.d/go.sh
export PATH=/usr/local/go/bin:$PATH
```

## 8. Clone source bằng root

```bash
cd /
git clone --branch main --single-branch https://github.com/Truonghocdot/bet.git /app
mkdir -p /app/bin /app/gin/storage/logs /app/gate/storage/logs
chown -R root:root /app
```

## 9. Tạo secret production

```bash
DB_PASSWORD="$(openssl rand -hex 24)"
AUTH_TOKEN_SECRET="$(openssl rand -hex 48)"
INTERNAL_TOKEN="$(openssl rand -hex 48)"
APP_KEY="base64:$(openssl rand -base64 32 | tr -d '\n')"
ADMIN_PHONE="0901234567"
ADMIN_PASSWORD="$(openssl rand -hex 12)"

cat >/root/fh88u-credentials <<EOF
DB_PASSWORD=$DB_PASSWORD
AUTH_TOKEN_SECRET=$AUTH_TOKEN_SECRET
INTERNAL_TOKEN=$INTERNAL_TOKEN
APP_KEY=$APP_KEY
ADMIN_PHONE=$ADMIN_PHONE
ADMIN_PASSWORD=$ADMIN_PASSWORD
EOF

chmod 600 /root/fh88u-credentials
source /root/fh88u-credentials
```

Không xóa file `/root/fh88u-credentials`; file này được dùng cho backup và lưu tài khoản admin ban đầu.

## 10. Bật PostgreSQL và Redis

```bash
systemctl enable --now postgresql redis-server cron supervisor nginx
```

Tạo PostgreSQL database:

```bash
source /root/fh88u-credentials

runuser -u postgres -- psql <<EOF
CREATE ROLE fh88u LOGIN PASSWORD '$DB_PASSWORD';
CREATE DATABASE fh88u OWNER fh88u ENCODING 'UTF8';
EOF
```

## 11. Tạo env cho Gin

```bash
source /root/fh88u-credentials

cat >/app/gin/.env <<EOF
APP_NAME=fh88u-gin
HTTP_ADDR=127.0.0.1:8081
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s
HTTP_SHUTDOWN_TIMEOUT=10s
DB_MAX_OPEN_CONNS=40
DB_MAX_IDLE_CONNS=20
DB_CONN_MAX_LIFETIME=30m
DB_CONN_MAX_IDLE_TIME=5m

CONTENT_ASSET_BASE_URL=https://admin.fh88u.win
POPUP_VIDEO_FILE_PATH=/app/vue/public/bg.mp4
DATABASE_URL=postgresql://fh88u:$DB_PASSWORD@127.0.0.1:5432/fh88u?sslmode=disable

AUTH_TOKEN_SECRET=$AUTH_TOKEN_SECRET
AUTH_TOKEN_TTL=2h
AUTH_REFRESH_TOKEN_TTL=720h
PUBLIC_REGISTER_URL=https://fh88u.win/register
CORS_ALLOWED_ORIGINS=https://fh88u.win,https://admin.fh88u.win,https://daily.fh88u.win

REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=2

# Chat global tắt mặc định. Bật đồng thời ở Gin, Laravel và Vue sau smoke test.
CHAT_GLOBAL_ENABLED=false
CHAT_ROOM_CODE=global

GATE_BASE_URL=http://127.0.0.1:8082
GIN_INTERNAL_TOKEN=$INTERNAL_TOKEN
GATE_INTERNAL_TOKEN=$INTERNAL_TOKEN
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

TCG_ENABLED=false
TCG_DEFAULT_LANGUAGE=VI
TCG_DEFAULT_GAME_MODE=1
TCG_DEFAULT_IP_ADDRESS=127.0.0.1
TCG_DEFAULT_WEB_PLATFORM=html5-desktop
TCG_DEFAULT_MOBILE_PLATFORM=html5
TCG_DEFAULT_LOTTERY_LOBBY_GAME_CODE=Lobby
TCG_LAUNCH_RETURN_URL=https://fh88u.win/play
TCG_DEFAULT_CURRENCY=VND
TCG_GAME_LIST_REDIS_KEY=shared:tcg:game-list:v1
TCG_PREVIEW_ZERO_BALANCE=false
EOF

chmod 600 /app/gin/.env
```

## 12. Tạo env cho Gate

```bash
source /root/fh88u-credentials

cat >/app/gate/.env <<EOF
APP_NAME=fh88u-gate
HTTP_ADDR=127.0.0.1:8082
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s
HTTP_SHUTDOWN_TIMEOUT=10s

GIN_INTERNAL_BASE_URL=http://127.0.0.1:8081
GIN_INTERNAL_TOKEN=$INTERNAL_TOKEN
GATE_INTERNAL_TOKEN=$INTERNAL_TOKEN

SHARED_REDIS_ADDR=127.0.0.1:6379
SHARED_REDIS_PASSWORD=
SHARED_REDIS_DB=2
EXCHANGE_RATE_REDIS_KEY=shared:exchange-rate:usdt-vnd
GATE_PUBLIC_BASE_URL=https://gate.fh88u.win

NOWPAYMENTS_BASE_URL=https://api.nowpayments.io
NOWPAYMENTS_API_KEY=
NOWPAYMENTS_IPN_SECRET=
NOWPAYMENTS_PRICE_CURRENCY=usd
NOWPAYMENTS_PAY_CURRENCY=USDTTRC20

TCG_ENABLED=false
TCG_BASE_URL=
TCG_HTTP_TIMEOUT=30s
TCG_MERCHANT_CODE=
TCG_MERCHANT_DES_KEY=
TCG_MERCHANT_SIGN_KEY=
TCG_REPORT_FTP_HOST=
TCG_REPORT_FTP_PORT=21
TCG_REPORT_FTP_USERNAME=
TCG_REPORT_FTP_PASSWORD=
TCG_REPORT_FTP_BASE_DIR=
TCG_GAME_LIST_SYNC_ENABLED=false
TCG_GAME_LIST_SYNC_INTERVAL=30s
TCG_GAME_LIST_REDIS_KEY=shared:tcg:game-list:v1
TCG_GAME_LIST_PLATFORM=all
TCG_GAME_LIST_CLIENT_TYPE=all
TCG_GAME_LIST_LANGUAGE=VI
TCG_GAME_LIST_PAGE=0
TCG_GAME_LIST_PAGE_SIZE=0
EOF

chmod 600 /app/gate/.env
```

## 13. Tạo env cho Laravel Admin

```bash
source /root/fh88u-credentials

cat >/app/admin/.env <<EOF
APP_NAME=FH88U
APP_ENV=production
APP_KEY=$APP_KEY
APP_DEBUG=false
APP_URL=https://admin.fh88u.win
APP_TIMEZONE=Asia/Ho_Chi_Minh
APP_LOCALE=vi
APP_FALLBACK_LOCALE=vi
APP_FAKER_LOCALE=vi_VN
APP_MAINTENANCE_DRIVER=file
BCRYPT_ROUNDS=12

HOST_MOBILE=https://fh88u.win
VUE_ADMIN_CONTROL_URL=https://fh88u.win/auth/sso

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
DB_PASSWORD=$DB_PASSWORD
DB_SSLMODE=disable

SESSION_DRIVER=redis
SESSION_LIFETIME=120
SESSION_ENCRYPT=true
SESSION_PATH=/
SESSION_DOMAIN=admin.fh88u.win
SESSION_SECURE_COOKIE=true
SESSION_HTTP_ONLY=true
SESSION_SAME_SITE=lax

BROADCAST_CONNECTION=log
FILESYSTEM_DISK=public
QUEUE_CONNECTION=redis
QUEUE_FAILED_DRIVER=null
CACHE_STORE=redis

REDIS_CLIENT=phpredis
REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379
REDIS_DB=0
REDIS_CACHE_CONNECTION=cache
REDIS_CACHE_DB=1
REDIS_SHARED_DB=2
REDIS_SHARED_PREFIX=

CHAT_GLOBAL_ENABLED=false
CHAT_ROOM_CODE=global

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

GIN_INTERNAL_TOKEN=$INTERNAL_TOKEN
TCG_GATE_BASE_URL=http://127.0.0.1:8082
TCG_GATE_INTERNAL_TOKEN=$INTERNAL_TOKEN
TCG_GATE_TIMEOUT=15
TCG_DEFAULT_CURRENCY=VND

MAIL_MAILER=log
MAIL_FROM_ADDRESS=hello@fh88u.win
MAIL_FROM_NAME=FH88U
EOF

chown root:www-data /app/admin/.env
chmod 640 /app/admin/.env
```

## 14. Tạo env cho hai frontend

```bash
cat >/app/vue/.env <<'EOF'
VITE_API_BASE_URL=https://api.fh88u.win
VITE_ALLOWED_HOSTS=fh88u.win
VITE_ENABLE_DEVTOOLS=false
VITE_CHAT_GLOBAL_ENABLED=false
VITE_CHAT_ROOM_CODE=global
DEV=false
EOF

cat >/app/agency/.env <<'EOF'
VITE_API_BASE_URL=https://api.fh88u.win
VITE_ALLOWED_HOSTS=daily.fh88u.win
DEV=false
EOF
```

## 15. Build Gin và Gate bằng root

```bash
export PATH=/usr/local/go/bin:$PATH

cd /app/gin
go mod download
go build -trimpath -o /app/bin/gin-api ./cmd/api
go build -trimpath -o /app/bin/gin-engine ./cmd/engine

cd /app/gate
go mod download
go build -trimpath -o /app/bin/gate ./cmd/webhooks
```

## 16. Build Vue và Agency bằng root

Chạy tuần tự để giảm RAM khi build:

```bash
cd /app/vue
pnpm install --frozen-lockfile
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vue-tsc --build
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vite build

cd /app/agency
pnpm install --frozen-lockfile
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vue-tsc --build
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vite build
```

## 17. Cài và build Laravel Admin bằng root

```bash
cd /app/admin
COMPOSER_ALLOW_SUPERUSER=1 composer install \
  --no-dev --prefer-dist --optimize-autoloader --no-progress
pnpm install --frozen-lockfile
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vite build
```

## 18. Migrate database và tạo admin

```bash
cd /app/admin
php artisan optimize:clear
php artisan migrate --force
php artisan db:seed --class=Database\\Seeders\\GameRoomSeeder --force
php artisan db:seed --class=Database\\Seeders\\ExchangeRateSettingSeeder --force
php artisan db:seed --class=Database\\Seeders\\VietQrBankSeeder --force
```

Chỉ seed ba class trên. Không chạy `php artisan db:seed`, vì seeder tổng có tài khoản/demo data dùng mật khẩu mẫu.

Tạo tài khoản super admin với mật khẩu đã sinh:

```bash
source /root/fh88u-credentials
cd /app/admin

FH88U_ADMIN_PHONE="$ADMIN_PHONE" FH88U_ADMIN_PASSWORD="$ADMIN_PASSWORD" \
php artisan tinker --execute='App\Models\User::query()->updateOrCreate(["phone" => getenv("FH88U_ADMIN_PHONE")], ["name" => "Super Admin", "email" => null, "password" => Illuminate\Support\Facades\Hash::make(getenv("FH88U_ADMIN_PASSWORD")), "role" => 0, "status" => 1, "phone_verified_at" => now()]);'
```

Cache Laravel và cấp quyền ghi cho PHP-FPM:

```bash
cd /app/admin
php artisan storage:link || true
php artisan config:cache
php artisan route:cache
php artisan view:cache
php artisan filament:optimize

chown -R root:www-data /app/admin/storage /app/admin/bootstrap/cache
chmod -R 775 /app/admin/storage /app/admin/bootstrap/cache
setfacl -R -m u:root:rwx -m u:www-data:rwx /app/admin/storage /app/admin/bootstrap/cache
find /app/admin/storage /app/admin/bootstrap/cache -type d \
  -exec setfacl -m d:u:root:rwx -m d:u:www-data:rwx {} +

systemctl reload php8.3-fpm
```

## 19. Cấu hình Supervisor chạy app bằng root

```bash
mkdir -p /var/log/fh88u
chmod 750 /var/log/fh88u

cat >/etc/supervisor/conf.d/fh88u.conf <<'EOF'
[program:fh88u-gin-api]
command=/app/bin/gin-api
directory=/app/gin
user=root
autostart=true
autorestart=true
startsecs=5
startretries=10
stopsignal=TERM
stopasgroup=true
killasgroup=true
environment=HOME="/root",USER="root"
stdout_logfile=/var/log/fh88u/gin-api.log
stdout_logfile_maxbytes=20MB
stdout_logfile_backups=10
redirect_stderr=true

[program:fh88u-gin-engine]
command=/app/bin/gin-engine
directory=/app/gin
user=root
autostart=true
autorestart=true
startsecs=5
startretries=10
stopsignal=TERM
stopasgroup=true
killasgroup=true
environment=HOME="/root",USER="root"
stdout_logfile=/var/log/fh88u/gin-engine.log
stdout_logfile_maxbytes=20MB
stdout_logfile_backups=10
redirect_stderr=true

[program:fh88u-gate]
command=/app/bin/gate
directory=/app/gate
user=root
autostart=true
autorestart=true
startsecs=5
startretries=10
stopsignal=TERM
stopasgroup=true
killasgroup=true
environment=HOME="/root",USER="root"
stdout_logfile=/var/log/fh88u/gate.log
stdout_logfile_maxbytes=20MB
stdout_logfile_backups=10
redirect_stderr=true

[program:fh88u-queue]
command=/usr/bin/php artisan queue:work redis --sleep=1 --tries=3 --timeout=120 --queue=default
directory=/app/admin
user=root
autostart=true
autorestart=true
startsecs=5
startretries=10
stopsignal=TERM
stopwaitsecs=180
stopasgroup=true
killasgroup=true
environment=HOME="/root",USER="root"
stdout_logfile=/var/log/fh88u/queue.log
stdout_logfile_maxbytes=20MB
stdout_logfile_backups=10
redirect_stderr=true

[program:fh88u-scheduler]
command=/usr/bin/php artisan schedule:work
directory=/app/admin
user=root
autostart=true
autorestart=true
startsecs=5
startretries=10
stopsignal=TERM
stopwaitsecs=30
stopasgroup=true
killasgroup=true
environment=HOME="/root",USER="root"
stdout_logfile=/var/log/fh88u/scheduler.log
stdout_logfile_maxbytes=20MB
stdout_logfile_backups=10
redirect_stderr=true

[group:fh88u]
programs=fh88u-gin-api,fh88u-gin-engine,fh88u-gate,fh88u-queue,fh88u-scheduler
priority=999
EOF

supervisorctl reread
supervisorctl update
supervisorctl restart 'fh88u:*'
```

## 20. Cấu hình IP thật khi dùng Cloudflare

```bash
{
  echo 'real_ip_header CF-Connecting-IP;'
  echo 'real_ip_recursive on;'
  curl -fsSL https://www.cloudflare.com/ips-v4 | sed 's#^#set_real_ip_from #; s#$#;#'
  curl -fsSL https://www.cloudflare.com/ips-v6 | sed 's#^#set_real_ip_from #; s#$#;#'
} >/etc/nginx/conf.d/cloudflare-realip.conf
```

## 21. Tạo Nginx HTTP để cấp SSL

```bash
mkdir -p /var/www/letsencrypt

cat >/etc/nginx/sites-available/fh88u.conf <<'EOF'
map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

server {
    listen 80;
    listen [::]:80;
    server_name fh88u.win;
    root /app/vue/dist;
    location ^~ /.well-known/acme-challenge/ { root /var/www/letsencrypt; }
    location / { try_files $uri $uri/ /index.html; }
}

server {
    listen 80;
    listen [::]:80;
    server_name daily.fh88u.win;
    root /app/agency/dist;
    location ^~ /.well-known/acme-challenge/ { root /var/www/letsencrypt; }
    location / { try_files $uri $uri/ /index.html; }
}

server {
    listen 80;
    listen [::]:80;
    server_name admin.fh88u.win;
    root /app/admin/public;
    index index.php index.html;
    location ^~ /.well-known/acme-challenge/ { root /var/www/letsencrypt; }
    location / { try_files $uri $uri/ /index.php?$query_string; }
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
    }
}

server {
    listen 80;
    listen [::]:80;
    server_name api.fh88u.win;
    location ^~ /.well-known/acme-challenge/ { root /var/www/letsencrypt; }
    location ^~ /internal/ { return 404; }
    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    listen [::]:80;
    server_name gate.fh88u.win;
    location ^~ /.well-known/acme-challenge/ { root /var/www/letsencrypt; }
    location = /healthz { proxy_pass http://127.0.0.1:8082; }
    location ^~ /v1/webhooks/deposits/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_request_buffering off;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location / { return 404; }
}
EOF

ln -sfn /etc/nginx/sites-available/fh88u.conf /etc/nginx/sites-enabled/fh88u.conf
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx
```

## 22. Mở firewall

Nếu SSH dùng cổng mặc định `22`:

```bash
ufw allow 22/tcp
ufw allow 'Nginx Full'
ufw --force enable
```

Nếu SSH dùng cổng khác, thay `22` bằng đúng cổng trước khi bật UFW.

## 23. Cấp SSL Let's Encrypt

```bash
certbot certonly --webroot -w /var/www/letsencrypt \
  --non-interactive --agree-tos --register-unsafely-without-email \
  --cert-name fh88u.win \
  -d fh88u.win \
  -d admin.fh88u.win \
  -d api.fh88u.win \
  -d gate.fh88u.win \
  -d daily.fh88u.win
```

## 24. Bật Nginx HTTPS production

```bash
cat >/etc/nginx/sites-available/fh88u.conf <<'EOF'
map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

server {
    listen 80;
    listen [::]:80;
    server_name fh88u.win admin.fh88u.win api.fh88u.win gate.fh88u.win daily.fh88u.win;
    location ^~ /.well-known/acme-challenge/ { root /var/www/letsencrypt; }
    location / { return 301 https://$host$request_uri; }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name fh88u.win;
    ssl_certificate /etc/letsencrypt/live/fh88u.win/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/fh88u.win/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    root /app/vue/dist;
    index index.html;
    client_max_body_size 20m;
    add_header X-Content-Type-Options nosniff always;
    add_header X-Frame-Options SAMEORIGIN always;
    location = /index.html {
        add_header Cache-Control "no-cache, no-store, must-revalidate" always;
        try_files $uri =404;
    }
    location / { try_files $uri $uri/ /index.html; }
    location ~* \.(css|js|png|jpe?g|gif|ico|svg|webp|woff2?|ttf|mp4)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        try_files $uri =404;
    }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name daily.fh88u.win;
    ssl_certificate /etc/letsencrypt/live/fh88u.win/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/fh88u.win/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    root /app/agency/dist;
    index index.html;
    add_header X-Content-Type-Options nosniff always;
    add_header X-Frame-Options SAMEORIGIN always;
    location = /index.html {
        add_header Cache-Control "no-cache, no-store, must-revalidate" always;
        try_files $uri =404;
    }
    location / { try_files $uri $uri/ /index.html; }
    location ~* \.(css|js|png|jpe?g|gif|ico|svg|webp|woff2?|ttf)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        try_files $uri =404;
    }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name admin.fh88u.win;
    ssl_certificate /etc/letsencrypt/live/fh88u.win/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/fh88u.win/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    root /app/admin/public;
    index index.php index.html;
    client_max_body_size 50m;
    add_header X-Content-Type-Options nosniff always;
    add_header X-Frame-Options SAMEORIGIN always;
    location / { try_files $uri $uri/ /index.php?$query_string; }
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        fastcgi_param DOCUMENT_ROOT $realpath_root;
        fastcgi_read_timeout 300;
    }
    location ~ /\.(?!well-known).* { deny all; }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.fh88u.win;
    ssl_certificate /etc/letsencrypt/live/fh88u.win/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/fh88u.win/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    client_max_body_size 20m;
    add_header X-Content-Type-Options nosniff always;
    location ^~ /internal/ { return 404; }
    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_read_timeout 300s;
    }
    location ^~ /v1/play/rooms/ {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 300s;
    }
    location = /v1/wallets/stream {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Connection '';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 300s;
    }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name gate.fh88u.win;
    ssl_certificate /etc/letsencrypt/live/fh88u.win/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/fh88u.win/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    client_max_body_size 20m;
    add_header X-Content-Type-Options nosniff always;
    location = /healthz {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location ^~ /v1/webhooks/deposits/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_request_buffering off;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    location / { return 404; }
}
EOF

nginx -t
systemctl reload nginx
```

Tạo hook reload Nginx sau khi Certbot tự renew:

```bash
mkdir -p /etc/letsencrypt/renewal-hooks/deploy

cat >/etc/letsencrypt/renewal-hooks/deploy/reload-nginx <<'EOF'
#!/usr/bin/env bash
set -e
/usr/sbin/nginx -t
/bin/systemctl reload nginx
EOF

chmod 755 /etc/letsencrypt/renewal-hooks/deploy/reload-nginx
```

## 25. Backup PostgreSQL mỗi ngày

```bash
mkdir -p /var/backups/fh88u/postgres

cat >/usr/local/sbin/fh88u-backup <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
source /root/fh88u-credentials
STAMP="$(date +%F-%H%M%S)"
export PGPASSWORD="$DB_PASSWORD"
pg_dump -h 127.0.0.1 -U fh88u -d fh88u \
  | gzip >"/var/backups/fh88u/postgres/fh88u-${STAMP}.sql.gz"
find /var/backups/fh88u/postgres -type f -name '*.sql.gz' -mtime +7 -delete
EOF

chmod 700 /usr/local/sbin/fh88u-backup
echo '15 2 * * * root /usr/local/sbin/fh88u-backup >>/var/log/fh88u-backup.log 2>&1' \
  >/etc/cron.d/fh88u-backup
chmod 644 /etc/cron.d/fh88u-backup
```

## 26. Kiểm tra cuối

```bash
supervisorctl status 'fh88u:*'
curl -s http://127.0.0.1:8081/healthz
curl -s http://127.0.0.1:8082/healthz
curl -I https://fh88u.win
curl -I https://admin.fh88u.win/up
curl -I https://api.fh88u.win/healthz
curl -I https://gate.fh88u.win/healthz
curl -I https://daily.fh88u.win
```

Xem tài khoản admin đã sinh:

```bash
cat /root/fh88u-credentials
```

Đăng nhập tại `https://admin.fh88u.win/admin` và đổi mật khẩu ngay.

## 27. Log và restart

```bash
supervisorctl status 'fh88u:*'
supervisorctl restart 'fh88u:*'
tail -f /var/log/fh88u/*.log
tail -f /app/admin/storage/logs/laravel.log
tail -f /var/log/nginx/error.log
```

## 28. Cập nhật source lần sau

```bash
cd /app
git pull --ff-only origin main

export PATH=/usr/local/go/bin:$PATH

cd /app/gin
go mod download
go build -trimpath -o /app/bin/gin-api ./cmd/api
go build -trimpath -o /app/bin/gin-engine ./cmd/engine

cd /app/gate
go mod download
go build -trimpath -o /app/bin/gate ./cmd/webhooks

cd /app/vue
pnpm install --frozen-lockfile
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vue-tsc --build
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vite build

cd /app/agency
pnpm install --frozen-lockfile
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vue-tsc --build
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vite build

cd /app/admin
COMPOSER_ALLOW_SUPERUSER=1 composer install \
  --no-dev --prefer-dist --optimize-autoloader --no-progress
pnpm install --frozen-lockfile
NODE_OPTIONS=--max-old-space-size=1536 ./node_modules/.bin/vite build
php artisan optimize:clear
php artisan migrate --force
php artisan config:cache
php artisan route:cache
php artisan view:cache
php artisan filament:optimize

systemctl reload php8.3-fpm
supervisorctl restart 'fh88u:*'
systemctl reload nginx
```

`TCG` và `NOWPayments` đang tắt/để trống theo mặc định. Sau khi có credential thật, sửa `/app/gin/.env` và `/app/gate/.env`, rồi chạy:

```bash
supervisorctl restart 'fh88u:*'
```

## 29. Bật Chat Global sau khi deploy

Chat chỉ được bật sau khi migration đã chạy và đã có bot profile/câu mẫu trong admin. Không tạo bot trong bảng `users`.

```bash
cd /app/admin
php artisan chat:import-templates /root/chat-templates.csv --dry-run
php artisan chat:import-templates /root/chat-templates.csv
php artisan chat:prune --message-days=30 --audit-days=90
```

Tạo 10-30 bot profile tại `https://admin.fh88u.win/admin`, kiểm tra tin nhắn và moderation trước. Sau smoke test hai tài khoản, đặt các biến sau thành `true` trong cả `/app/gin/.env`, `/app/admin/.env` và `/app/vue/.env`:

```dotenv
CHAT_GLOBAL_ENABLED=true
VITE_CHAT_GLOBAL_ENABLED=true
CHAT_ROOM_CODE=global
VITE_CHAT_ROOM_CODE=global
```

Build lại Vue và restart process:

```bash
cd /app/vue
pnpm build

supervisorctl restart fh88u-gin-api fh88u-queue fh88u-scheduler
```

Rollback không xóa dữ liệu: đặt hai feature flag về `false`, build lại Vue, rồi restart ba process trên. Tin nhắn giữ tối đa 30 ngày; audit moderation giữ 90 ngày.
