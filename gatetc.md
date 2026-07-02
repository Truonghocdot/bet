# Tóm tắt tài liệu TC-Gaming

Ngày tổng hợp: 2026-07-02

## Nguồn tài liệu

- Common API: https://doc.tc-gaming.com/TW/
- Reports: https://doc.tc-gaming.com/REPORT/
- Appendix: https://doc.tc-gaming.com/APPENDIX/

## 1. Thống kê nhanh

- Bộ tài liệu được chia thành 3 phần chính: `Common API`, `Reports`, `Appendix`.
- `Common API` có 10 mục hướng dẫn, nhưng thực tế bao phủ 12 method code chính:
  `cm`, `up`, `kom`, `gb`, `ft`, `ftoa`, `cs`, `lg`, `tgl`, `gml`, `gfl`, `gtl`.
- `Reports` được chia thành 7 nhóm nghiệp vụ:
  Lottery Common, TLOTTO, ELOTTO/TCG SEA, RNG/FISH, LIVE, PVP, SPORT.
- `Reports` kết hợp 2 kiểu lấy dữ liệu:
  gọi API HTTP trực tiếp và lấy file batch qua FTP.
- `Appendix` là phần tra cứu lớn, gồm:
  mã tiền tệ, product type, game type, ngôn ngữ, error code, page code, game icon, MIF form, quy tắc thêm report, fund transfer exception.

## 2. Bản chất tích hợp

TC-Gaming trong bộ docs này đang mô tả mô hình `transfer wallet`.

Luồng tổng quát:

1. Merchant tạo user trên TC-Gaming.
2. Merchant đồng bộ mật khẩu nếu cần.
3. Merchant truy vấn số dư theo `product_type`.
4. Merchant chuyển tiền vào hoặc ra ví game bằng API transfer.
5. Merchant launch game bằng `lg`.
6. Merchant đối soát giao dịch và cược chơi bằng report API hoặc FTP.

## 3. Giao thức và bảo mật request

- Gateway request trong docs dùng `POST` vào endpoint dạng `doBusiness.do`.
- `Content-Type`: `application/x-www-form-urlencoded`.
- Request có 3 field chính:
  `merchant_code`, `params`, `sign`.
- `params` là JSON được:
  `DES-ECB-PKCS5Padding` rồi `Base64`.
- `sign` là:
  `SHA-256(params + merchant_key/hash_key)`.
- Docs nhấn mạnh phải dùng `UTF-8`.
- Docs khuyến nghị timeout API từ `30s` trở lên.
- Docs cũng cảnh báo không nên build contract quá cứng, vì vendor có thể thêm field mới mà không coi là breaking change.
- Có sample code cho Java, PHP, C# và VB.NET.

## 4. Danh sách Common API

| Method | Chức năng | Ghi chú thực tế |
| --- | --- | --- |
| `cm` | Tạo / đăng ký player | Gần như idempotent, nếu user đã tồn tại có thể trả thành công thay vì lỗi. |
| `up` | Cập nhật mật khẩu player | Docs nói rõ sẽ đồng bộ xuống game provider, nếu fail có backoffice retry. |
| `kom` | Kick out lotto member | Phục vụ logout hoặc kick khỏi hệ thống lottery. |
| `gb` | Lấy số dư | Cần `username` và `product_type`. |
| `ft` | Chuyển tiền vào/ra | Dùng `fund_type`, `amount`, `reference_no`. |
| `ftoa` | Rút toàn bộ tiền khỏi product wallet | Thường dùng để sweep balance về merchant wallet. |
| `cs` | Kiểm tra trạng thái giao dịch | Tra theo `ref_no` hoặc `reference_no`. |
| `lg` | Launch game | Trả `game_url`, là API quan trọng nhất phía frontend/game entry. |
| `tgl` | Lấy danh sách game tổng | Có filter `product_type`, `client_type`, `game_type`, `platform`, paging. |
| `gml` | TCG Live model list | Dùng cho hệ sinh thái TCG Live. |
| `gfl` | TCG Live fighter list | Dùng cho hệ sinh thái TCG Live. |
| `gtl` | TCG Live tournament list | Dùng cho hệ sinh thái TCG Live. |

## 5. Tham số bắt buộc cần truyền

### 5.1 Request ngoài cùng

Mọi request Common API đều cần 3 field ngoài cùng:

- `merchant_code`: mã merchant được cấp
- `params`: chuỗi JSON đã mã hóa bằng DES
- `sign`: `SHA256(params + merchant_hash_key)`

### 5.2 `cm` - Create / Register Player

Tham số bắt buộc trong `params`:

- `method = "cm"`
- `username`
- `password`

Tham số không bắt buộc:

- `currency`

### 5.3 `up` - Update Player Password

Tham số bắt buộc:

- `method = "up"`
- `username`
- `password`

### 5.4 `kom` - Kick Out Member

Tham số bắt buộc:

- `method = "kom"`
- `username`

### 5.5 `gb` - Get Balance

Tham số bắt buộc:

- `method = "gb"`
- `username`
- `product_type`

### 5.6 `ft` - Fund Transfer

Tham số bắt buộc:

- `method = "ft"`
- `username`
- `product_type`
- `fund_type`
- `amount`
- `reference_no`

Ghi chú:

- `fund_type = 1` là nạp vào game
- `fund_type = 2` là rút ra khỏi game

### 5.7 `ftoa` - Fund Transfer Out All

Tham số bắt buộc:

- `method = "ftoa"`
- `username`
- `product_type`
- `reference_no`

### 5.8 `cs` - Check Transfer Status

Tham số bắt buộc:

- `method = "cs"`
- `product_type`
- `ref_no`

### 5.9 `lg` - Launch Game

Theo bảng parameter trong docs, các field bắt buộc là:

- `method = "lg"`
- `username`
- `product_type`
- `ip_address`
- `platform`
- `game_mode`
- `game_code`

Tham số không bắt buộc nhưng dùng rất nhiều:

- `nickname`
- `vip_level`
- `view`
- `back_url`
- `lottery_bet_mode`
- `language`
- `series`

Ghi chú quan trọng:

- Docs đánh dấu `ip_address` là bắt buộc, nhưng ngay trong mô tả lại ghi nếu không truyền thì hệ thống mặc định `127.0.0.1`.
- `platform` được đánh dấu bắt buộc, nhưng mô tả cũng ghi riêng `TCG SEA` có thể tự nhận diện mobile/web.
- `series` được docs ghi là bắt buộc cho một số lottery product type như `2`, `384`, `420`.

### 5.10 `tgl` - TCG Game List

Tham số bắt buộc:

- `method = "tgl"`
- `product_type`
- `platform`
- `client_type`
- `game_type`

Tham số không bắt buộc:

- `language`
- `page`
- `page_size`

### 5.11 `gml` - TCG Live Model List

Theo đúng bảng docs, tham số bắt buộc:

- `method = "gml"`
- `merchant_code`

Tham số không bắt buộc:

- `language`
- `active`

Ghi chú:

- Đây là điểm hơi lạ vì request ngoài cùng đã có `merchant_code`, nhưng bảng của docs vẫn liệt kê thêm `merchant_code` bên trong `params`.

### 5.12 `gfl` - TCG Live Fighter List

Tham số bắt buộc:

- `method = "gfl"`
- `language`
- `username`

Tham số không bắt buộc:

- `isAll`

### 5.13 `gtl` - TCG Live Tournament List

Tham số bắt buộc:

- `method = "gtl"`
- `language`

## 6. Common API thuần HTTP: ý nghĩa và cách gọi

12 method dưới đây là nhóm `Common API` gọi trực tiếp qua HTTP, không liên quan đến FTP. Chúng phục vụ các thao tác runtime như tạo user, chuyển tiền, launch game, lấy danh sách game và lấy dữ liệu TCG Live.

### 6.1 Cách gọi “thuần” chung

Mọi Common API đều đi theo cùng một pattern:

1. Tạo JSON `params` gốc.
2. Mã hóa `params` bằng DES.
3. Tạo `sign = SHA256(params + merchant_hash_key)`.
4. Gửi `POST` tới `doBusiness.do` với form fields:
   `merchant_code`, `params`, `sign`.

Mẫu gọi tối thiểu:

```text
POST /doBusiness.do
Content-Type: application/x-www-form-urlencoded

merchant_code=YOUR_MERCHANT_CODE
params=ENCRYPTED_DES_BASE64_STRING
sign=SHA256_HEX_STRING
```

Nếu nói theo nghĩa “thuần”, phần dưới đây là JSON `params` trước khi mã hóa.

### 6.2 `cm` - Create / Register Player

Ý nghĩa:

- Tạo player phía TC-Gaming.
- Nếu user đã tồn tại, docs cho thấy API này có thể hoạt động gần giống “confirm player exists” hơn là chỉ thuần create.

Khi nào dùng:

- Khi user của hệ thống mình lần đầu cần đi vào hệ sinh thái game TC-Gaming.

JSON `params` mẫu:

```json
{
  "method": "cm",
  "username": "phoenix",
  "password": "1q2w3e4r",
  "currency": "CNY"
}
```

Response chính:

- `status`
- `error_desc`

### 6.3 `up` - Update Player Password

Ý nghĩa:

- Đồng bộ lại mật khẩu player lên TC-Gaming.

Khi nào dùng:

- Khi người dùng đổi mật khẩu ở hệ thống chính và bạn muốn đồng bộ sang phía vendor.

JSON `params` mẫu:

```json
{
  "method": "up",
  "username": "phoenix",
  "password": "1q2w3e4r"
}
```

Response chính:

- `status`
- `error_desc`

### 6.4 `kom` - Kick Out Member

Ý nghĩa:

- Kick user ra khỏi phiên lottery/vendor session hiện tại.

Khi nào dùng:

- Khi cần force logout khỏi game hoặc làm sạch session trước khi relaunch.

JSON `params` mẫu:

```json
{
  "method": "kom",
  "username": "phoenix"
}
```

Response chính:

- `status`
- `error_desc`

### 6.5 `gb` - Get Balance

Ý nghĩa:

- Lấy số dư của player trong một `product_type` cụ thể.

Khi nào dùng:

- Trước khi transfer out
- Khi đồng bộ số dư ví game
- Khi hiển thị số dư product wallet riêng

JSON `params` mẫu:

```json
{
  "method": "gb",
  "username": "phoenix",
  "product_type": 7
}
```

Response chính:

- `status`
- `balance`
- `error_desc`

### 6.6 `ft` - Fund Transfer

Ý nghĩa:

- Chuyển tiền giữa ví merchant và ví game của một `product_type`.

Khi nào dùng:

- Người chơi bấm vào game, cần nạp tiền vào ví game
- Người chơi rời game, cần rút tiền từ ví game về ví chính

JSON `params` mẫu:

```json
{
  "method": "ft",
  "username": "phoenix",
  "product_type": 7,
  "fund_type": "1",
  "amount": 100.5,
  "reference_no": "TCGTESTFI2017001"
}
```

Diễn giải:

- `fund_type = 1`: chuyển vào game
- `fund_type = 2`: chuyển ra khỏi game
- `reference_no` phải unique để retry và đối soát

Response chính:

- `status`
- `error_desc`

### 6.7 `ftoa` - Fund Transfer Out All

Ý nghĩa:

- Rút toàn bộ số dư còn lại của player trong product wallet về lại hệ thống merchant.

Khi nào dùng:

- Khi người chơi thoát game
- Khi cần sweep sạch ví game trước một luồng đóng phiên/chuyển sản phẩm

JSON `params` mẫu:

```json
{
  "method": "ftoa",
  "username": "phoenix",
  "product_type": 7,
  "reference_no": "TCGTESTFI2017001"
}
```

Response chính:

- `status`
- `error_desc`

### 6.8 `cs` - Check Transfer Status

Ý nghĩa:

- Kiểm tra trạng thái của giao dịch transfer trước đó.

Khi nào dùng:

- Khi `ft` hoặc `ftoa` timeout
- Khi network lỗi và không chắc giao dịch đã thực sự thành công hay chưa

JSON `params` mẫu:

```json
{
  "method": "cs",
  "product_type": 7,
  "ref_no": "TCGTESTFI2017001"
}
```

Response chính:

- `status`
- `transfer_status` hoặc trạng thái nghiệp vụ tương đương
- `error_desc`

Lưu ý:

- Theo docs, nếu trạng thái là failed thì nên gọi lại transfer mới với `reference_no` mới.
- Nếu trạng thái là unknown thì nên xem là ca cần đối soát kỹ hoặc liên hệ vendor.

### 6.9 `lg` - Launch Game

Ý nghĩa:

- Trả về `game_url` để đưa người dùng vào game hoặc lobby.

Khi nào dùng:

- Khi user bấm vào game/lobby bất kỳ thuộc TC-Gaming.

JSON `params` mẫu tối thiểu:

```json
{
  "method": "lg",
  "username": "phoenix",
  "product_type": 4,
  "game_code": "A00070",
  "game_mode": "1",
  "language": "CN",
  "platform": "html5-desktop"
}
```

JSON `params` mẫu cho lottery lobby:

```json
{
  "method": "lg",
  "username": "phoenix",
  "product_type": 2,
  "game_code": "Lobby",
  "game_mode": "1",
  "platform": "html5-desktop",
  "lottery_bet_mode": "Traditional",
  "view": "Lobby",
  "series": [
    {
      "game_group_code": "SSC",
      "prize_mode_id": 1,
      "max_series": 1980,
      "min_series": 1700,
      "max_bet_series": 1980,
      "default_series": 1700
    }
  ]
}
```

Response chính:

- `status`
- `error_desc`
- `game_url`
- đôi khi có thêm `pt_username`, `pt_password`

Lưu ý:

- Đây là API nhiều biến thể nhất.
- `view`, `lottery_bet_mode`, `series` thường quan trọng với lottery.
- `game_code` có thể là lobby hoặc một game cụ thể.

### 6.10 `tgl` - TCG Game List

Ý nghĩa:

- Lấy danh sách game thuộc một product/game type.

Khi nào dùng:

- Đồng bộ catalog game
- Build menu game phía frontend/backoffice
- Mapping `game_code` trước khi launch

JSON `params` mẫu:

```json
{
  "platform": "all",
  "page": 0,
  "method": "tgl",
  "product_type": "7",
  "client_type": "all",
  "game_type": "RNG",
  "page_size": 0
}
```

Response chính:

- `status`
- `games`
- `page_info`
- `error_desc`

Giải thích từng field của mỗi object trong `games[]`:

- `displayStatus`
  Trạng thái hiển thị của game trong catalog do TCG trả về. Thực tế thường thấy `0`, nên nên coi đây là cờ trạng thái/phân loại hiển thị từ phía vendor.

- `gameType`
  Nhóm game cấp cao của item.
  Ví dụ thường gặp:
  `RNG`, `LIVE`, `FISH`, `PVP`, `SPORTS`, `ELOTT`.
  Field này dùng để lọc catalog và cũng phản ánh giá trị `game_type` khi gọi `tgl`.

- `gameName`
  Tên game để hiển thị ra UI.
  Đây là label người dùng nhìn thấy, có thể thay đổi theo ngôn ngữ hoặc merchant.

- `tcgGameCode`
  Mã game nội bộ của TCG.
  Đây là field rất quan trọng vì khi launch game bằng `lg`, thường bạn cần truyền `game_code` tương ứng từ catalog này.

- `productCode`
  Mã viết tắt của provider/product.
  Ví dụ như `PT`, `AG`, `GG`.
  Field này map với cột `Abbrev` trong bảng `Product Type` của appendix.

- `productType`
  Mã product dưới dạng string trong response, ví dụ `"3"`, `"4"`, `"7"`.
  Đây là version string của `product_type` mà request đã gọi vào.
  Muốn hiểu mã này là gì thì phải đối chiếu bảng `Product Type` ở appendix.

- `platform`
  Danh sách nền tảng game hỗ trợ.
  Ví dụ:
  `flash`, `html5`, `html5-desktop`, hoặc chuỗi ghép như `flash,html5`.
  Field này rất hữu ích để quyết định client web/mobile có launch được game hay không.

- `gameSubType`
  Loại con của game trong từng `gameType`.
  Với `RNG`, docs appendix có bảng mapping như:
  `JP`, `SC`, `SM`, `TG`, `VP`, `AC`.
  Với `LIVE`, docs appendix có các subtype như:
  `BAC`, `SIC`, `ROU`, `DAT`, `BJ`, `NN`...
  Đây là field tốt để nhóm game ở UI.

- `trialSupport`
  Cho biết game có hỗ trợ chế độ thử hay không.
  Nếu `true`, game có thể dùng được khi launch bằng `game_mode = 0`.
  Nếu `false`, nên coi như chỉ dùng cho tài khoản thật.

Ghi chú:

- Trong log thực tế của merchant mình còn thấy field `showIcon`, dù sample docs tóm tắt không luôn liệt kê nó. Đây là URL icon game do TCG trả thêm.
- `productType` trong response là string, nhưng trong request `tgl` thì `product_type` là int.

Mapping field với docs:

- `productType` <-> Appendix `Product Type`
- `gameSubType` <-> Appendix `Game List`
- `tcgGameCode` -> thường dùng tiếp cho `lg`

### 6.11 `gml` - TCG Live Model List

Ý nghĩa:

- Lấy danh sách model/room trong hệ sinh thái TCG Live.

Khi nào dùng:

- Build trang lobby/live room list
- Đồng bộ metadata TCG Live

JSON `params` mẫu:

```json
{
  "method": "gml",
  "merchant_code": "YOUR_MERCHANT_CODE",
  "language": "EN",
  "active": 0
}
```

Response chính:

- `status`
- `details`
- `error_desc`

Trong `details` có thể thấy các field như:

- `id`
- `name`
- `active`
- `streaming`
- `gameCode`
- `gameName`

### 6.12 `gfl` - TCG Live Fighter List

Ý nghĩa:

- Lấy danh sách fighter/focus list liên quan đến user trong TCG Live.

Khi nào dùng:

- Build danh sách focus/online fighter cho user
- Hiển thị danh sách nhân vật/đối tượng live đang theo dõi

JSON `params` mẫu:

```json
{
  "method": "gfl",
  "language": "EN",
  "username": "phoenix",
  "isAll": false
}
```

Response chính:

- `status`
- `details`
- `error_desc`

### 6.13 `gtl` - TCG Live Tournament List

Ý nghĩa:

- Lấy danh sách tournament/tag/game set của TCG Live.

Khi nào dùng:

- Build menu tournament/category trong phần live

JSON `params` mẫu:

```json
{
  "method": "gtl",
  "language": "EN"
}
```

Response chính:

- `status`
- `details`
- `error_desc`

## 7. Lưu ý quan trọng theo từng API

### 7.1 `cm` - Create/Register Player

- Username theo docs có ràng buộc format và độ dài.
- Password được yêu cầu độ dài tối thiểu và tối đa.
- Currency được truyền lúc tạo user.
- Nên xem `cm` là điểm khởi tạo mapping giữa user merchant và user TC-Gaming.

### 7.2 `ft` / `ftoa` / `cs`

- `reference_no` phải unique, đây là trụ cột cho retry và đối soát.
- `ft` có `fund_type` để phân biệt chuyển vào hay chuyển ra.
- Amount thông thường dùng 2 số lẻ; docs có ghi lottery có trường hợp hỗ trợ 4 số lẻ.
- Appendix có mục `Fund Transfer Exception`, nên đọc kỹ trước khi implement retry/job đối soát.
- `cs` nên được dùng làm API xác minh kết quả khi `ft` hoặc `ftoa` gặp timeout hay response mơ hồ.

### 7.3 `lg` - Launch Game

Đây là phần có nhiều biến thể nhất trong docs.

Thông số thường gặp:

- `product_type`
- `game_code`
- `game_mode`
- `platform`
- `language`
- `view`
- `lottery_bet_mode`
- `series`
- `return_url`

Lưu ý:

- Cùng một `product_type` có thể có nhiều kiểu launch khác nhau giữa web và mobile.
- Lottery launch cần nhiều config hơn RNG/LIVE.
- `series` được dùng để setup prize mode / odds / default series cho từng `game_group_code`.
- Docs có nhắc đến `MIF Form URL`, tức là có những flow launch cần form trung gian.

## 8. Product type và nghiệp vụ đáng chú ý

Từ appendix và các ví dụ launch/report, những nhóm product hay gặp gồm:

- `2`: TCG lottery
- `384`: TCG SEA lottery
- `420`: TCG Vietnam lottery / Vietnam-related flows
- `460`: TCG Live

Ghi chú:

- Appendix có cột cho biết product nào là `TRANSFER`, product nào là `SEAMLESS`, và một số product hỗ trợ cả hai.
- `460 / TCG Live` trong appendix có ghi chú `Pause Integration`, nên nếu muốn làm TCG Live thì cần xác minh lại với vendor trước.

### 8.1 Lưu ý về link appendix

- Link đúng của bảng `Language Code` là:
  `https://doc.tc-gaming.com/APPENDIX/#appendix_LanguageCode`
- Link đúng của bảng `Product Type` là:
  `https://doc.tc-gaming.com/APPENDIX/#appendix_ProductType`

Nếu mục tiêu là tra `product_type` để gọi `gb`, `ft`, `lg`, `tgl` thì phải xem bảng `Product Type`, không phải bảng `Language Code`.

### 8.2 Bảng Product Type rút gọn cho tích hợp hiện tại

Đây là các dòng quan trọng rút ra từ appendix, đủ để dùng cho các luồng đang triển khai:

| Product Type | Abbrev | Mô tả | Wallet | Loại game hỗ trợ |
| --- | --- | --- | --- | --- |
| `2` | `TCG LOTTO` | TCG LOTTO | `TRANSFER / SEAMLESS` | `LOTT` |
| `3` | `PT` | Playtech | `TRANSFER` | `LIVE`, `RNG`, `FISH` |
| `4` | `AG` | PlayAce | `TRANSFER / SEAMLESS` | `LIVE`, `RNG`, `FISH` |
| `384` | `TCG_SEA` | TCG_SEA Đông Nam Á | `TRANSFER / SEAMLESS` | `ELOTT` |
| `420` | `TCG_LOTTO_VN` | TCG_LOTTO_VN Việt Nam | `TRANSFER / SEAMLESS` | `LOTT` |
| `460` | `TCG_LIVE` | TCG Live (tạm dừng tích hợp) | `TRANSFER / SEAMLESS` | `LOTT` theo bảng appendix, nhưng docs có ghi chú pause |

### 8.3 Kết luận thực tế sau khi test `tgl`

Từ log probe thực tế với merchant hiện tại:

- `product_type=7` trả `status=0` nhưng `games=null`
- `product_type=2` trả `status=0` nhưng `games=null`
- `product_type=420` trả `status=0` nhưng `games=null`
- `product_type=384` trả `status=1` với lỗi hệ thống
- `product_type=460` trả `status=1` với lỗi hệ thống
- `product_type=3` trả game list thật

Điều này cho thấy ít nhất với merchant hiện tại, nếu gọi `tgl` để lấy game RNG thì `product_type=3` là lựa chọn đang hoạt động đúng.

Ví dụ request thực tế đang hoạt động:

```json
{
  "method": "tgl",
  "product_type": 3,
  "platform": "all",
  "client_type": "all",
  "game_type": "RNG",
  "language": "VI",
  "page": 1,
  "page_size": 100
}
```

Ví dụ response thực tế:

- `status = 0`
- `games` có dữ liệu
- các phần tử gồm:
  `gameName`, `tcgGameCode`, `productCode`, `productType`, `platform`, `gameSubType`, `showIcon`, `trialSupport`

## 9. Phần Reports được tổ chức như thế nào

Report docs được chia làm 2 kiểu:

### 9.1 Query trực tiếp qua API

Dùng cho:

- Tra cứu lịch sử cược theo user
- Lấy draw result / winner history
- Lấy round details
- Lấy unsettled data trong một số sản phẩm

### 9.2 Batch qua FTP

Dùng cho:

- Settled bet details
- Cancelled / corrected data ở một số sản phẩm
- Dữ liệu đối soát lớn để đồng bộ về merchant

Docs ghi rõ:

- file batch được cập nhật theo chu kỳ `5 phút` hoặc `15 phút` tùy nhóm sản phẩm
- file được giữ trên FTP trong `15 ngày`
- naming và folder cần được ingest tự động, không nên parse thủ công

## 10. Thống kê report theo nhóm

### 10.1 Lottery Common

Những method code nhìn thấy rõ trong docs:

- `glgl`: get game list
- `gldr`: get lottery draw result
- `gwnh`: get winner note history
- `glrd`: get lottery round details
- `glmt`: get live member transaction report

### 10.2 TLOTTO

- `glmoh`: member order history
- `gluoh`: unsettled batch/API theo `batch_name`
- Ngoài ra còn có các luồng settled / cancelled qua FTP

### 10.3 ELOTTO / TCG SEA

- `elmbd`: bet detail by member
- `elubd`: unsettled bet detail
- settled data lấy qua FTP

### 10.4 RNG / FISH

- `bdm`: bet detail by member
- settled data lấy qua FTP

### 10.5 LIVE

- `lbdm`: live bet detail by member
- settled data lấy qua FTP

### 10.6 SPORT

- `spmbd`: sports bet detail by member
- `spubd`: sports unsettled bet detail
- settled data lấy qua FTP

### 10.7 PVP

- Docs có mục FTP riêng cho PVP bet details
- Phần này nên xem như reconciliation source theo batch

## 11. Giới hạn và pattern thường gặp trong report

- Nhiều API report theo member giới hạn date range tối đa `7 ngày`.
- Có endpoint cho phép `page_size` lớn hơn, ví dụ `5000`.
- Report theo member phù hợp cho backoffice tra cứu.
- Report FTP phù hợp cho ETL/job đồng bộ và đối soát.
- Response schema không hoàn toàn đồng nhất:
  có endpoint trả `status`, có endpoint trả `success`, có endpoint trả `error_desc`.

## 12. Appendix - những phần nên đọc kỹ

### 12.1 API return code

Appendix có bảng mã trả về. Nhìn nhanh:

- `0`: thành công
- các mã khác bao gồm:
  merchant sai/không tồn tại, request sai format, decrypt fail, sai method, user không tồn tại, số dư không đủ, product không hỗ trợ, transfer đang pending...

Khuyến nghị:

- Centralize mapping error code ngay từ đầu.
- Tách `business error` và `transport error`.

### 12.2 Product type / game type / currency / language

Appendix có bảng mapping cho:

- `product_type`
- `game_type`
- `currency`
- `language`
- `playcode / language code`

Nếu team frontend/backend cần hiển thị label thân thiện, nên build lookup table riêng từ appendix thay vì hard-code trong source.

### 12.3 Lottery page / view / mode

Appendix và ví dụ launch có nhiều page code như:

- `Lobby`
- `Game_List`
- `Betting`
- `Draw_History`
- `Orders`
- `Trend_Chart`

Điều này quan trọng nếu cần deep-link vào đúng màn hình lottery.

### 12.4 MIF Form URL

Docs có nói đến MIF form và trang appendix có section riêng. Nếu sau này frontend launch game qua form redirect thay vì direct URL thì cần đọc phần này kỹ.

## 13. Những điểm dễ implement sai

- Build DTO response quá cứng, đến khi vendor thêm field thì vỡ parser.
- Không quản lý idempotency cho `reference_no`.
- Không dùng `cs` để verify transfer sau timeout.
- Không tách rõ product `TRANSFER` và `SEAMLESS`.
- Không có job ingest FTP tự động, dẫn đến mất dữ liệu sau 15 ngày.
- Đồng bộ report mà không thống nhất timezone.
- Lập kế hoạch cho `TCG Live` dù appendix đang đánh dấu pause.

## 14. Đề xuất thứ tự implement

1. Viết helper encrypt/decrypt/signature cho `doBusiness.do`.
2. Làm xong `cm`, `gb`, `ft`, `ftoa`, `cs`.
3. Hoàn thiện `lg` và `tgl`, sau đó mapping `product_type` + `game_code`.
4. Đọc appendix để tạo lookup table nội bộ.
5. Làm report ingestion:
   query theo member trước, FTP batch sau.
6. Bổ sung đối soát transfer exception và alert cho giao dịch pending.

## 15. Kết luận

TC-Gaming là bộ docs khá đầy, nhưng có 3 đặc điểm cần nhớ:

- API giao dịch và launch game là phần dễ tích hợp nhanh nhất.
- Report/FTP mới là phần quyết định chất lượng đối soát và vận hành.
- Appendix rất quan trọng, vì nhiều giá trị `product_type`, `page code`, `language`, `error code` không nên đoán theo cảm tính.

Nếu làm tích hợp thật, mình sẽ ưu tiên:

- đồng bộ transfer wallet core
- chuẩn hóa launch game
- sau cùng mới đồng bộ report và automation đối soát
