1. Giới thiệu tổng quát

NOWPayments là cổng thanh toán non-custodial (không giữ tiền hộ) hỗ trợ hơn 300+ cryptocurrencies, trong đó có USDT TRC20 (rất phù hợp cho site bet vì phí rẻ, tốc độ nhanh).
Cho phép nhận thanh toán crypto → tự động convert sang đồng bạn muốn (ví dụ: USDT TRC20) và chuyển thẳng vào ví payout của bạn.
Rất phù hợp với ngành iGaming / Betting (có tài liệu riêng cho betting/casino).

Base URL: https://api.nowpayments.io 
2. Authentication

Đăng ký tại nowpayments.io → thêm payout wallet (ví nhận tiền).
Tạo API Key (dùng để gọi API).
Tạo IPN Secret Key (chỉ hiện 1 lần khi tạo → phải lưu lại ngay). Dùng để verify webhook.
Cách auth: Thêm header x-api-key: YOUR_API_KEY vào mọi request.

3. Luồng tích hợp chính cho site bet (Deposit)

User chọn nạp tiền (số tiền USD hoặc tương đương).
Server gọi API Create Payment → nhận về:
payment_id
Địa chỉ ví deposit (hoặc link invoice)
pay_currency (ví dụ: usdttrc20)

Hiển thị QR code + địa chỉ ví cho user nạp.
User chuyển USDT TRC20 vào địa chỉ đó.
NOWPayments gửi Webhook (IPN) về ipn_callback_url của bạn khi trạng thái thay đổi.
Kiểm tra trạng thái qua GET Payment Status nếu cần.

Endpoint quan trọng nhất:

POST /v1/payment → Tạo payment/invoice
Tham số quan trọng: price_amount, price_currency (thường là usd), pay_currency (usdttrc20), ipn_callback_url, success_url, order_id (tùy chọn, dùng để map với user của bạn).

4. Webhook (IPN) – Rất quan trọng cho realtime

Đây là cách tốt nhất để biết user đã nạp thành công hay chưa (không nên poll liên tục).
NOWPayments sẽ POST về URL bạn cung cấp khi payment status thay đổi (waiting, confirming, finished, failed, partially_paid…).
Phải verify signature bằng HMAC-SHA512 với IPN Secret Key để tránh giả mạo.
Webhook payload chứa: payment_id, payment_status, pay_amount, pay_currency, outcome_amount, fee, v.v.

Lưu ý quan trọng:

Có thể nhận webhook nhiều lần (repeated deposit) → kiểm tra parent_payment_id.
Sai mạng (sai asset) → mặc định cần xử lý thủ công (có thể bật auto convert).

5. Các endpoint quan trọng khác

NhómEndpoint chínhMục đích chính cho site betCurrenciesGET available currenciesLấy danh sách coin hỗ trợMinimum amountGET minimum payment amountKiểm tra số tiền nạp tối thiểuEstimate priceGET estimate priceƯớc tính số coin cần nạpPayment StatusGET /v1/payment/{payment_id}Kiểm tra trạng thái paymentMass PayoutsMass payout endpointsRút tiền hàng loạt (withdraw cho user)Webhookipn_callback_urlNhận thông báo realtime
