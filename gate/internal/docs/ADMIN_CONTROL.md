# Nghiệp vụ: Hệ thống Điều khiển kết quả (Admin Control)

## 1. Mục tiêu
Đảm bảo lợi nhuận cho nhà cái bằng cách can thiệp vào kết quả của các kỳ xổ (Win Go, K3, 5D Lô tô). Hệ thống hỗ trợ cả chế độ điều khiển thủ công và tự động.

## 2. Các thành phần chính
- **Bet Monitor**: Theo dõi tổng tiền cược realtime của từng cửa trong kỳ hiện tại.
- **Manual Override**: Admin chọn số/màu cụ thể sẽ ra.
- **Auto-Kill Config**: Hệ thống tự động chọn cửa có tổng tiền cược thấp nhất để mở thưởng.

## 3. Quy trình thực hiện (Business Flow)

### A. Chế độ Thủ công (Manual)
1. Admin truy cập màn hình **Quản lý kỳ xổ**.
2. Hệ thống hiển thị:
   - Thời gian còn lại của kỳ.
   - Thống kê cược: Ví dụ Win Go có 100tr đặt Lớn, 20tr đặt Nhỏ.
3. Admin quyết định kỳ này phải ra **Nhỏ** để tối ưu tiền thắng cho sàn.
4. Admin nhập kết quả mong muốn và nhấn "Lưu".
5. Khi đồng hồ về 0, Backend kiểm tra lệnh từ Admin và trả về kết quả đã được chỉ định thay vì kết quả ngẫu nhiên.

### B. Chế độ Tự động (Auto-Smart)
1. Admin bật cấu hình "Tự động cân cửa" (Auto-Kill).
2. Backend trước khi mở thưởng 2 giây sẽ tính toán:
   - Cửa A: Tổng thanh toán (Payout) nếu thắng là 200tr.
   - Cửa B: Tổng thanh toán nếu thắng là 50tr.
3. Hệ thống tự động chọn kết quả thuộc cửa B.

## 4. Yêu cầu kỹ thuật (Backend)
- **Database/Redis**: Lưu trữ `planned_result` cho từng `period_no`.
- **Worker**: Service xử lý mở thưởng phải ưu tiên kiểm tra dữ liệu can thiệp trước khi gọi hàm Random.
- **Logging**: Mọi hành động can thiệp của Admin phải được log lại (Admin ID, IP, kết quả cũ, kết quả mới).

---

# Nghiệp vụ: Quy trình Rút tiền (Withdrawal)

## 1. Luồng nghiệp vụ
1. User yêu cầu rút tiền về địa chỉ ví USDT (TRC20).
2. Hệ thống tạm trừ số dư trong ví User (Status: PENDING).
3. Kiểm tra điều kiện (Risk Management):
   - Tổng cược (Turnover) đã đạt yêu cầu chưa? (Thường là x1 hoặc x2 số tiền nạp).
   - Có dấu hiệu gian lận, spam lệnh không?
4. Phê duyệt:
   - **Tự động**: Với các lệnh nhỏ (ví dụ < 100$).
   - **Thủ công**: Admin duyệt cho các lệnh lớn.
5. Thực hiện chuyển tiền qua API NOWPayments (Mass Payouts).
6. Cập nhật trạng thái thành SUCCESS và gửi thông báo cho User.

## 2. Các trạng thái lệnh rút
- `PENDING`: Chờ xử lý.
- `REVIEWING`: Đang kiểm tra rủi ro.
- `PROCESSING`: Đang gửi lệnh lên cổng thanh toán.
- `SUCCESS`: Hoàn thành.
- `REJECTED`: Từ chối (Hoàn tiền về ví User).
