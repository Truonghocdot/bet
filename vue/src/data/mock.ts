export type GameCard = {
  code: string
  title: string
  subtitle: string
  accent: string
  symbol: string
  enabled?: boolean
}

export type HomeWinningItem = {
  displayName: string
  description: string
  amount: string
  tag: string
  avatarUrl?: string
}

export type NotificationPreview = {
  title: string
  unreadCount: number
}

export const featuredGames: GameCard[] = [
  { code: 'wingo', title: 'Win Go', subtitle: 'Dự đoán màu sắc, nhận thưởng lớn mỗi phút', accent: '#ff6d66', symbol: 'rocket_launch' },
  { code: 'k3', title: 'K3', subtitle: 'Xúc xắc may mắn, tỉ lệ thắng cực cao', accent: '#e64545', symbol: 'casino' },
  { code: 'lottery', title: '5D Lotre', subtitle: 'Chọn số trúng vàng, vinh quang gõ cửa', accent: '#f6c32d', symbol: 'looks_5' },
  { code: 'trx_win', title: 'Trx Win', subtitle: 'Khai thác tiền số, bùng nổ lợi nhuận', accent: '#24b561', symbol: 'currency_bitcoin', enabled: false },
]

export const quickCategories = [
  { title: 'Xổ số', symbol: 'confirmation_number' },
  { title: 'Casino', symbol: 'casino' },
  { title: 'Bắn cá', symbol: 'skull' },
  { title: 'Thể thao', symbol: 'sports_soccer' },
  { title: 'Game bài', symbol: 'playing_cards' },
]

export const homeWinningFeed: HomeWinningItem[] = [
  { displayName: 'User***821', description: 'Vừa rút tiền thành công', amount: '+25,000,000đ', tag: 'Win Go 1m' },
  { displayName: 'Linh***9x', description: 'K3 Sicbo Master', amount: '+8,400,000đ', tag: 'K3 Lotre' },
  { displayName: 'Tuan***_pro', description: 'Jackpot Thể Thao', amount: '+102,000,000đ', tag: 'SABA Sport' },
]

export const accountNotifications: NotificationPreview = {
  title: 'Thông báo',
  unreadCount: 3,
}

export const gameVariants = [
  { label: '30 Giây', key: '30s', active: true },
  { label: '1 Phút', key: '1m', active: false },
  { label: '3 Phút', key: '3m', active: false },
  { label: '5 Phút', key: '5m', active: false },
]

export const gameHistoryRows = [
  { period: '...341', number: '2', size: 'Nhỏ', color: 'Đỏ' },
  { period: '...340', number: '7', size: 'Lớn', color: 'Xanh' },
  { period: '...339', number: '5', size: 'Lớn', color: 'Tím' },
  { period: '...338', number: '4', size: 'Nhỏ', color: 'Đỏ' },
]

export const promotionTabs = ['Tất cả', 'Thành viên mới', 'Hoàn trả', 'VIP Club']

export const promotionCards = [
  {
    title: 'Đặc quyền VIP Membership',
    description: 'Nâng cấp level, nhận quà sinh nhật và hoàn trả không giới hạn mỗi tuần.',
    tag: 'HOT',
    icon: 'star',
    accent: 'primary',
    primaryAction: 'Tham gia ngay',
  },
  {
    title: 'Hoàn trả 1.5% Casino',
    description: 'Tự động kết toán mỗi ngày.',
    tag: 'HOT',
    icon: 'casino',
    accent: 'secondary',
    primaryAction: 'Nhận ngay',
  },
  {
    title: 'Cược thể thao 0đ',
    description: 'Tặng vé cược miễn phí hàng tuần.',
    tag: 'NEW',
    icon: 'sports_soccer',
    accent: 'secondary',
    primaryAction: 'Nhận ngay',
  },
]

export const promotionTerms = [
  'Các chương trình khuyến mãi chỉ áp dụng cho một tài khoản duy nhất trên mỗi địa chỉ IP hoặc thiết bị.',
  'FF789 có quyền thay đổi hoặc chấm dứt khuyến mãi mà không cần thông báo trước trong trường hợp gian lận.',
]
