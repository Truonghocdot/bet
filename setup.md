# FH88U Production Setup

Hướng dẫn đầy đủ, tách từng khối command cho Ubuntu 22.04 nằm trong [`full-setup.md`](full-setup.md).

Cấu hình hiện tại:

- Source và build dùng `root`, thư mục `/app` thuộc `root`.
- Process ứng dụng chạy bằng Supervisor với `user=root`.
- Không tạo hoặc sử dụng user `deploy`.
- Nginx/PHP-FPM, PostgreSQL và Redis dùng daemon mặc định của Ubuntu.
- `fh88u.win`: app khách hàng.
- `admin.fh88u.win`: Laravel admin.
- `api.fh88u.win`: Gin API.
- `gate.fh88u.win`: webhook Gate.
- `daily.fh88u.win`: agency dashboard.

Lệnh vận hành chính:

```bash
supervisorctl status 'fh88u:*'
supervisorctl restart 'fh88u:*'
cat /root/fh88u-credentials
```
