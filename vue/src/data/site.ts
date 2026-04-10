export type GameRoom = {
  code: string
  title: string
  subtitle: string
  accent: string
  symbol: string
  category: string
  roundTime: string
  onlinePlayers: number
  minBet: string
  payout: string
  jackpot: string
  status: 'OPEN' | 'CLOSED' | 'COMING_SOON'
  featured?: boolean
}

export type HomeMetric = {
  title: string
  value: string
  description: string
  icon: string
  accent: string
}

export type HomeActivity = {
  title: string
  subtitle: string
  amount: string
  tag: string
  symbol: string
  tone: 'success' | 'info' | 'warning'
}

export type NewsArticle = {
  slug: string
  title: string
  excerpt: string
  cover: string
  category: string
  publishedAt: string
  readTime: string
  author: string
  tags: string[]
  content: string[]
  featured?: boolean
}

export type NotificationItem = {
  id: number
  title: string
  body: string
  category: string
  createdAt: string
  unread: boolean
  tone: 'info' | 'success' | 'warning'
  relatedSlug?: string
}

export const homeMetrics: HomeMetric[] = [
  {
    title: 'Ví VND',
    value: 'Đang đồng bộ',
    description: 'Số dư sẽ lấy từ backend khi có dữ liệu ví.',
    icon: 'account_balance_wallet',
    accent: '#004edb',
  },
  {
    title: 'Tài khoản mời',
    value: 'REF FF789A1',
    description: 'Mã giới thiệu hiện tại của bạn.',
    icon: 'group_add',
    accent: '#6c5a00',
  },
  {
    title: 'Thông báo mới',
    value: '04',
    description: 'Thông báo hệ thống chưa đọc.',
    icon: 'notifications',
    accent: '#b71211',
  },
  {
    title: 'Phòng đang mở',
    value: '03 trò',
    description: 'Win Go, K3 và Lô tô đang sẵn sàng.',
    icon: 'casino',
    accent: '#7e9cff',
  },
]

export const quickActions = [
  { title: 'Nạp tiền', symbol: 'add_card', to: '/deposit', accent: '#004edb' },
  { title: 'Thông báo', symbol: 'notifications', to: '/notifications', accent: '#b71211' },
  { title: 'Hoạt động', symbol: 'campaign', to: '/promotion', accent: '#6c5a00' },
  { title: 'Phòng chơi', symbol: 'sports_esports', to: '/play/wingo', accent: '#7e9cff' },
  { title: 'Cá nhân', symbol: 'person', to: '/account', accent: '#004edb' },
]

export const gameCategories = ['Tất cả', 'Win Go', 'K3', 'Lô tô', 'Sắp mở']

export const gameRooms: GameRoom[] = [
  {
    code: 'wingo',
    title: 'Win Go',
    subtitle: 'Phòng 30 giây, nhịp nhanh, tỷ lệ vào lệnh cao.',
    accent: '#004edb',
    symbol: 'rocket_launch',
    category: 'Win Go',
    roundTime: '30 giây',
    onlinePlayers: 1834,
    minBet: '10.000đ',
    payout: '98.6%',
    jackpot: '125.000.000đ',
    status: 'OPEN',
    featured: true,
  },
  {
    code: 'k3',
    title: 'K3',
    subtitle: 'Xúc xắc 3 số, chốt theo kỳ và kết quả realtime.',
    accent: '#b71211',
    symbol: 'casino',
    category: 'K3',
    roundTime: '1 phút',
    onlinePlayers: 1021,
    minBet: '20.000đ',
    payout: '97.9%',
    jackpot: '80.000.000đ',
    status: 'OPEN',
    featured: true,
  },
  {
    code: 'lottery',
    title: '5D Lô tô',
    subtitle: 'Chọn bộ số, canh nhịp quay, chốt trong từng kỳ.',
    accent: '#6c5a00',
    symbol: 'looks_5',
    category: 'Lô tô',
    roundTime: '3 phút',
    onlinePlayers: 856,
    minBet: '50.000đ',
    payout: '96.8%',
    jackpot: '250.000.000đ',
    status: 'OPEN',
    featured: true,
  },
  {
    code: 'trx_win',
    title: 'Trx Win',
    subtitle: 'Màn mở rộng hệ sinh thái, hiện đang trong giai đoạn chuẩn bị.',
    accent: '#7e9cff',
    symbol: 'currency_bitcoin',
    category: 'Sắp mở',
    roundTime: '--',
    onlinePlayers: 0,
    minBet: '--',
    payout: '--',
    jackpot: '--',
    status: 'COMING_SOON',
  },
]

export const homeActivities: HomeActivity[] = [
  { title: 'Rút thưởng thành công', subtitle: 'User***821 vừa rút tiền về tài khoản', amount: '+25,000,000đ', tag: 'Win Go 1m', symbol: 'payments', tone: 'success' },
  { title: 'K3 nổ lớn', subtitle: 'Linh***9x chốt trúng một ván rất đẹp', amount: '+8,400,000đ', tag: 'K3 Sicbo', symbol: 'casino', tone: 'info' },
  { title: 'Jackpot bất ngờ', subtitle: 'Tuan***_pro nhận thưởng trong phiên chiều', amount: '+102,000,000đ', tag: 'Lô tô', symbol: 'verified', tone: 'warning' },
]

export const newsArticles: NewsArticle[] = [
  {
    slug: 'thuong-nap-lan-dau-100-phan-tram',
    title: 'Thưởng nạp lần đầu lên đến 100%',
    excerpt: 'Người chơi mới nạp lần đầu sẽ nhận thêm giá trị khuyến mãi theo mốc đăng ký.',
    cover: 'from-primary via-primary-dim to-primary-container',
    category: 'Khuyến mãi',
    publishedAt: '2026-04-11T08:15:00+07:00',
    readTime: '2 phút',
    author: 'Hệ thống FF789',
    tags: ['newbie', 'bonus', 'deposit'],
    featured: true,
    content: [
      'Chương trình thưởng nạp lần đầu được thiết kế để người chơi làm quen với hệ thống và có thêm ngân sách trải nghiệm các phòng Win Go, K3 và Lô tô.',
      'Mỗi giao dịch hợp lệ sẽ được ghi nhận qua hệ thống giao dịch và trạng thái sẽ hiển thị trong hồ sơ người chơi.',
      'Trạng thái, điều kiện và thời gian áp dụng có thể thay đổi theo từng đợt vận hành.',
    ],
  },
  {
    slug: 'cap-nhat-phong-wingo-30s',
    title: 'Cập nhật Win Go 30s và tối ưu trải nghiệm',
    excerpt: 'Phòng Win Go được tinh chỉnh để vào lệnh nhanh hơn, giảm độ trễ giao diện.',
    cover: 'from-[#004edb] via-[#0058bb] to-[#7e9cff]',
    category: 'Tin hệ thống',
    publishedAt: '2026-04-10T22:05:00+07:00',
    readTime: '3 phút',
    author: 'Ban vận hành',
    tags: ['wingo', 'performance', 'room'],
    content: [
      'Win Go tiếp tục là phòng có nhịp đặt lệnh nhanh nhất trong hệ thống, được tối ưu cho các thao tác mobile.',
      'Trang phòng chơi hiện hiển thị danh sách game rõ ràng hơn để người chơi chọn đúng loại phòng trước khi join.',
      'Khi kết nối realtime sẵn sàng, các trạng thái kỳ và kết quả sẽ được đẩy trực tiếp về giao diện phòng.',
    ],
  },
  {
    slug: 'uu-dai-hoan-tra-theo-cap-do',
    title: 'Hoàn trả theo cấp độ thành viên',
    excerpt: 'Hệ thống ưu đãi được chia theo VIP để người chơi dễ theo dõi và nhận thưởng.',
    cover: 'from-[#6c5a00] via-[#fdd404] to-[#fff3b0]',
    category: 'VIP Club',
    publishedAt: '2026-04-09T13:30:00+07:00',
    readTime: '2 phút',
    author: 'Hệ thống FF789',
    tags: ['vip', 'cashback', 'reward'],
    content: [
      'Khi đạt mốc VIP, người chơi sẽ mở thêm các quyền lợi hoàn trả và ưu đãi riêng.',
      'Các bài tin trong mục Hoạt động sẽ ưu tiên cập nhật những chính sách hoàn trả và sự kiện ảnh hưởng trực tiếp đến người chơi.',
    ],
  },
  {
    slug: 'dang-ky-co-ma-gioi-thieu',
    title: 'Đăng ký bằng mã giới thiệu để mở quyền thưởng sớm',
    excerpt: 'Luồng đăng ký mới hỗ trợ mã giới thiệu, giúp user được gắn ref ngay từ lần đầu.',
    cover: 'from-[#b71211] via-[#ff7f7f] to-[#ffd6d6]',
    category: 'Referral',
    publishedAt: '2026-04-08T19:25:00+07:00',
    readTime: '2 phút',
    author: 'Ban vận hành',
    tags: ['referral', 'register', 'affiliate'],
    content: [
      'Mã giới thiệu được dùng ngay trong màn đăng ký để gắn ref code cho tài khoản mới.',
      'Mục tiêu là giúp người giới thiệu theo dõi ngay từ đầu và người được mời dễ nhận các ưu đãi phù hợp.',
    ],
  },
]

export const notificationItems: NotificationItem[] = [
  {
    id: 1,
    title: 'Nạp tiền thành công',
    body: 'Giao dịch nạp của bạn đã được xác nhận và cộng ví.',
    category: 'Ví',
    createdAt: '2026-04-11T07:40:00+07:00',
    unread: true,
    tone: 'success',
    relatedSlug: 'thuong-nap-lan-dau-100-phan-tram',
  },
  {
    id: 2,
    title: 'Bài tin mới đã được đăng',
    body: 'FF789 vừa cập nhật chương trình thưởng nạp lần đầu.',
    category: 'Tin tức',
    createdAt: '2026-04-11T06:55:00+07:00',
    unread: true,
    tone: 'info',
    relatedSlug: 'thuong-nap-lan-dau-100-phan-tram',
  },
  {
    id: 3,
    title: 'Hoàn trả VIP đã ghi nhận',
    body: 'Chu kỳ hoàn trả mới đã được cộng vào hồ sơ của bạn.',
    category: 'VIP',
    createdAt: '2026-04-10T22:20:00+07:00',
    unread: false,
    tone: 'warning',
    relatedSlug: 'uu-dai-hoan-tra-theo-cap-do',
  },
  {
    id: 4,
    title: 'Nhắc nhở bảo mật',
    body: 'Hãy bật xác thực hai lớp để bảo vệ tài khoản tốt hơn.',
    category: 'Bảo mật',
    createdAt: '2026-04-10T09:15:00+07:00',
    unread: false,
    tone: 'info',
  },
]

export function getNewsBySlug(slug: string) {
  return newsArticles.find((article) => article.slug === slug)
}

export function getRelatedNews(currentSlug: string, limit = 3) {
  return newsArticles.filter((article) => article.slug !== currentSlug).slice(0, limit)
}

export function getUnreadCount() {
  return notificationItems.filter((item) => item.unread).length
}

