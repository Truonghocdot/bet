# End User Register

## Mục tiêu

Cho phép người chơi tự tạo tài khoản bằng form đăng ký công khai.

Luồng này chỉ xử lý các dữ liệu tối thiểu cần thiết để tạo một account hợp lệ.

Chi tiết các bản ghi phải insert được tách riêng ở [Bản ghi cần insert](./insert-records.md).

Mỗi user khi đăng ký sẽ có mã và link referral riêng để chia sẻ cho người khác. Khi ấn vào link này thì redirect đến màn đăng ký và tự động điền mã referral vào form. Nếu người dùng đăng ký thành công thì sẽ được ghi nhận là một referral của người đã chia sẻ link.

Khi đăng ký không cần verify email hoặc phone ngay lập tức. Nhưng khi cần change/reset password hoặc các luồng nghiệp vụ quan trọng khác thì hệ thống sẽ yêu cầu verify email hoặc phone theo chính sách của hệ thống. Nếu người dùng không verify được thì sẽ không thể thực hiện các luồng nghiệp vụ quan trọng đó. Tuy nhiên, người dùng vẫn có thể đăng nhập và sử dụng các tính năng cơ bản của hệ thống nếu chưa verify email hoặc phone, tùy theo chính sách của hệ thống.

Khi đăng ký sẽ tiến hành insert user và insert các relation liên quan bằng transaction để đảm bảo tính nhất quán của dữ liệu. Nếu có lỗi trong quá trình insert thì sẽ rollback transaction và trả về lỗi cho người dùng.

## Input

- `name`
- `email`
- `phone` nếu hệ thống cho phép
- `password`
- `ref_code` hoặc `ref_link` nếu có affiliate

## UX liên quan (màn Auth)

Thiết kế UI auth (đăng nhập) có 2 tab:

- đăng nhập bằng `Số điện thoại`
- đăng nhập bằng `Email`

Hệ quả nghiệp vụ:

- `phone` là field optional, nhưng nếu user đã đăng ký bằng phone thì login có thể dùng phone.
- nếu user chỉ đăng ký bằng email thì tab phone phải hiển thị lỗi phù hợp (không leak enumeration ở luồng quên mật khẩu).

Ngoài ra UI có:

- `Nhớ mật khẩu`: chỉ là behavior phía client (lưu token/credential theo policy), backend không cần field riêng.
- `Quên mật khẩu?`: gọi luồng OTP reset password theo tài liệu auth-security.

## Validation

- `name` bắt buộc, độ dài hợp lệ
- `email` bắt buộc và duy nhất
- `phone` nếu có thì duy nhất
- `password` bắt buộc và đủ mạnh theo policy
- dữ liệu affiliate nếu có phải trỏ tới mã hợp lệ

## Side Effects

- tạo record trong `users`
- tạo `wallet` mặc định nếu hệ thống yêu cầu
- nếu có affiliate hợp lệ thì tạo/ghi nhận quan hệ referral
- không cho người dùng set các field nội bộ của ERP

## Không được phép từ màn end user

- không set `role`
- không set `status` tùy ý theo ERP
- không set `approved_by`
- không set `last_login_at`
- không tạo được các record kỹ thuật như `ledger`, `transaction`, `withdrawal request`

## Liên kết: Quên mật khẩu (OTP)

UI có entry `Quên mật khẩu?`.

Nghiệp vụ bắt buộc:

- public response phải generic, không để lộ email/phone có tồn tại hay không.
- OTP reset password chỉ lưu hash (không lưu OTP plaintext).
- có rate-limit/cooldown để chống spam.

Chi tiết xem: `admin/docs/auth-security/forgot-password-and-anti-spam.md`.

## Kết quả mong muốn

- user được tạo thành công
- account có thể login theo chính sách xác minh của hệ thống
- các luồng nghiệp vụ sau đó sẽ do hệ thống hoặc ERP quản lý tiếp
