# Tài liệu vận hành - Chỉnh sửa tài chính thủ công trong Filament

Tài liệu này dành cho quản trị viên vận hành hệ thống `admin` khi cần chỉnh tay dữ liệu tài chính vừa được bổ sung trong Filament.

Phạm vi tính năng:

- sửa `amount` và `timestamp` của giao dịch trong bảng `Giao dịch`
- sửa `amount` và `timestamp` của yêu cầu rút trong bảng `Yêu cầu rút`
- chỉnh trực tiếp số dư ví khả dụng của người dùng trong form hồ sơ user

Mục tiêu:

- cho phép đội vận hành xử lý sai lệch dữ liệu nhanh ngay trên ERP
- vẫn giữ dấu vết thay đổi qua `wallet_ledger_entries` khi có tác động tới ví
- tách rõ đâu là số dư khả dụng, đâu là số dư khóa để tránh thao tác nhầm

## 1. Sửa giao dịch trong bảng `Giao dịch`

Màn hình:

- Filament `Tài chính -> Giao dịch`
- action trên từng dòng: `Sửa số tiền`

Cho phép sửa:

- `Số tiền`
- `Thời gian tạo`
- `Thời gian duyệt`

Quy tắc nghiệp vụ:

- khi sửa `amount`, hệ thống tự tính lại `net_amount = amount - fee`
- nếu giao dịch là `DEPOSIT` và đã ở trạng thái `COMPLETED`:
  - ví người dùng sẽ được cộng hoặc trừ đúng phần chênh lệch
  - hệ thống ghi thêm một dòng `wallet_ledger_entries`
- nếu giao dịch chưa `COMPLETED`:
  - chỉ cập nhật dữ liệu trên record giao dịch
  - không tự cộng/trừ ví
- nếu là giao dịch đã hoàn tất nhưng không phải `DEPOSIT`:
  - hệ thống chặn sửa `amount` để tránh lệch sổ ví

Ví dụ:

- giao dịch nạp đã duyệt từ `1.000.000` sửa thành `1.200.000`
  - ví khả dụng tăng thêm `200.000`
- giao dịch nạp đã duyệt từ `1.000.000` sửa thành `900.000`
  - ví khả dụng giảm `100.000`

Lưu ý vận hành:

- chỉ sửa `amount` sau khi đã kiểm tra chắc chắn tiền thực tế nhận được
- nếu chỉ cần chỉnh mốc thời gian, giữ nguyên `amount`

## 2. Sửa yêu cầu rút trong bảng `Yêu cầu rút`

Màn hình:

- Filament `Tài chính -> Yêu cầu rút`
- action trên từng dòng: `Sửa số tiền`

Cho phép sửa:

- `Số tiền`
- `Thời gian tạo`
- `Thời gian duyệt`
- `Thời gian chi trả`

Modal có thêm thông tin:

- `Số dư đang bị đóng băng`

Quy tắc nghiệp vụ:

- khi sửa `amount`, hệ thống tự tính lại `net_amount = amount - fee`
- khi sửa `amount`, hệ thống chỉ đồng bộ `locked_balance` của ví gắn với lệnh rút
- `balance` khả dụng không đổi trong luồng chỉnh tay này
- nếu `amount` tăng:
  - `locked_balance` tăng đúng phần chênh lệch
- nếu `amount` giảm:
  - `locked_balance` giảm đúng phần chênh lệch
- nếu sau khi tính toán mà `locked_balance` âm:
  - hệ thống chặn lưu
- với lệnh có trạng thái `REJECTED`, `CANCELED`, `PAID`:
  - admin vẫn có thể chỉnh timestamp
  - nếu đổi `amount` thì hệ thống sẽ chặn

Ví dụ:

- lệnh rút từ `500.000` sửa thành `650.000`
  - `locked_balance` tăng `150.000`
  - `balance` giữ nguyên
- lệnh rút từ `500.000` sửa thành `300.000`
  - `locked_balance` giảm `200.000`
  - `balance` giữ nguyên

Lưu ý rất quan trọng:

- đây là luồng chỉnh số tiền bị khóa theo lệnh rút, không phải luồng hoàn tiền
- nếu mục tiêu là trả tiền lại cho người dùng, hãy dùng đúng nghiệp vụ `Từ chối` lệnh rút thay vì sửa `amount`

## 3. Chỉnh số dư ví trong hồ sơ người dùng

Màn hình:

- Filament `Người dùng`
- vào chi tiết user bất kỳ
- section mới: `Số dư ví`

Cho phép chỉnh:

- `Số dư ví VND`
- `Số dư ví USDT`

Quy tắc nghiệp vụ:

- đây là chỉnh trực tiếp `balance` khả dụng
- `locked_balance` không thay đổi trong luồng này
- nếu ví chưa tồn tại, hệ thống tự tạo ví mới tương ứng unit
- khi có thay đổi số dư:
  - hệ thống ghi `wallet_ledger_entries`
  - note sẽ ghi rõ đây là điều chỉnh từ form người dùng bởi admin thao tác
- ví được đưa về `ACTIVE` sau khi lưu

Khi nên dùng:

- cần đồng bộ số dư thực tế sau đối soát thủ công
- cần cấp bù hoặc trừ bù nhanh cho user mà không đi qua luồng nạp/rút

Khi không nên dùng:

- không dùng để mô phỏng lệnh rút
- không dùng để thay thế cho duyệt nạp khi giao dịch nạp đã tồn tại

## 4. Checklist thao tác an toàn cho admin

Trước khi sửa:

- xác nhận đúng user
- xác nhận đúng đơn vị ví `VND` hoặc `USDT`
- xác nhận cần sửa `amount`, `timestamp`, hay cả hai
- với lệnh rút, xác nhận rõ đang cần đổi số tiền bị khóa chứ không phải hoàn tiền

Sau khi sửa:

- reload lại dòng dữ liệu vừa chỉnh
- kiểm tra số dư ví trên hồ sơ user
- nếu có tác động tới ví, kiểm tra `wallet_ledger_entries`
- nếu là lệnh rút, kiểm tra lại `locked_balance` đã đổi đúng phần chênh lệch

## 5. Một số tình huống thường gặp

### 5.1 Nạp tiền thực nhận khác số đã tạo

Xử lý:

- vào `Tài chính -> Giao dịch`
- chọn đúng lệnh nạp
- bấm `Sửa số tiền`
- nhập số tiền thực nhận đúng

Kết quả:

- nếu lệnh đã `COMPLETED`, ví sẽ tự bù/trừ phần chênh lệch

### 5.2 Lệnh rút tạo sai amount nhưng chưa chi trả

Xử lý:

- vào `Tài chính -> Yêu cầu rút`
- bấm `Sửa số tiền`
- nhập lại `amount` đúng

Kết quả:

- `locked_balance` đổi theo chênh lệch
- `balance` không đổi

### 5.3 Cần chỉnh tay số dư user

Xử lý:

- vào hồ sơ user
- sửa ở section `Số dư ví`
- lưu form

Kết quả:

- số dư khả dụng đổi ngay
- có ledger để truy vết

## 6. Ghi chú cho đội vận hành

- ưu tiên dùng đúng workflow nghiệp vụ trước, chỉ chỉnh tay khi thật sự cần
- với các lệnh đã `PAID` hoặc đã hoàn tất nghiệp vụ, không cố sửa `amount` nếu hệ thống chặn
- mọi thay đổi số dư nên được đối chiếu lại với log nội bộ hoặc bằng chứng giao dịch thực tế
