# Auth Security

Tài liệu này mô tả nghiệp vụ bảo mật cho luồng auth end-user trong `gin` trước khi triển khai code.

Phạm vi hiện tại:

- chống spam / chống abuse cho `register`, `login`, `forgot password`
- quên mật khẩu bằng OTP qua `email` hoặc `phone`
- cách `gin` phối hợp với `gate` để gửi thông báo
- các bản ghi và cache key cần có để triển khai an toàn

File chính:

- [forgot-password-and-anti-spam.md](./forgot-password-and-anti-spam.md)

