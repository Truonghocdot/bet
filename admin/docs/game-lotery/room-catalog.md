# Room Catalog (Hard-code)

## Mục tiêu

- mỗi tab time ở UI tương ứng đúng 1 room.
- room code ổn định để API/engine/ERP cùng dùng.

## Danh sách room

### Wingo

- `wingo_30s`: duration `30`
- `wingo_1m`: duration `60`
- `wingo_3m`: duration `180`
- `wingo_5m`: duration `300`

### K3

- `k3_1m`: duration `60`
- `k3_3m`: duration `180`
- `k3_5m`: duration `300`
- `k3_10m`: duration `600`

### Lottery (5D)

- `lottery_1m`: duration `60`
- `lottery_3m`: duration `180`
- `lottery_5m`: duration `300`
- `lottery_10m`: duration `600`

## Cấu hình DB dự kiến (`game_rooms`)

- `code`: unique
- `game_type`: tinyint (`WINGO|K3|LOTTERY`)
- `duration_seconds`: int
- `bet_cutoff_seconds`: int, mặc định `5`
- `status`: tinyint (`ACTIVE|INACTIVE`)
- `sort_order`: int

## Rule vận hành

- `code` không đổi sau khi phát hành.
- room `INACTIVE` không mở kỳ mới.
- room `ACTIVE` luôn phải có kỳ kế tiếp do engine đảm bảo.
