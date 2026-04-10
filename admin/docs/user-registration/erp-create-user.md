# ERP Create User

## Mục tiêu

Cho phép quản trị viên tạo user thủ công trong ERP.

Luồng này dành cho backoffice, không phải self-service.

Khi create ở đây cũng phải tiến hành insert user và insert các relation liên quan bằng transaction để đảm bảo tính nhất quán của dữ liệu. Nếu có lỗi trong quá trình insert thì sẽ rollback transaction và trả về lỗi cho người dùng.

Chi tiết các bản ghi phải insert được tách riêng ở [Bản ghi cần insert](./insert-records.md).

## Input

ERP được phép nhập thêm các field quản trị:

- `name`
- `email`
- `phone`
- `password`
- `role`
- `status`
- `email_verified_at`
- `phone_verified_at`
- `last_login_at`

Ngoài ra form create ERP còn có nhóm provisioning để chủ động tạo record liên quan:

- `provision_wallets`
- `provision_vnd_wallet`
- `provision_affiliate_profile`
- `affiliate_status`
- `provision_account_withdrawal_info`
- `withdrawal_unit`
- `withdrawal_provider_code`
- `withdrawal_account_name`
- `withdrawal_account_number`
- `withdrawal_is_default`

## Validation

- `email` bắt buộc và duy nhất
- `phone` nếu có thì duy nhất
- `role` phải nằm trong enum hệ thống
- `status` phải nằm trong enum hệ thống
- mật khẩu phải được hash theo chuẩn của model

## Side Effects

- tạo record `users`
- có thể tạo kèm các dữ liệu liên quan nếu nghiệp vụ yêu cầu:
  - `wallet` (`VND`, `USDT`) nếu bật provisioning
  - `affiliate_profile` nếu bật provisioning, trong đó `ref_code` tự sinh duy nhất và `ref_link` tự tạo từ `ref_code`
  - `account_withdrawal_info` nếu bật provisioning
- ghi nhận audit user tạo

## Nghiệp vụ cần lưu ý

- ERP create không nên dùng cùng form với end user register
- đây là luồng vận hành nội bộ, nên giữ đầy đủ quyền kiểm soát
- nếu tạo user thủ công xong thì trạng thái mặc định phải được quyết định rõ ngay từ policy hệ thống

## Kết quả mong muốn

- admin tạo được user nhanh
- có thể backfill dữ liệu thiếu cho account cũ
- dễ audit và kiểm soát quyền hơn so với luồng register công khai
