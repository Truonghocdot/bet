export type PlayResult = {
  periodNo: string
  result: string
  bigSmall: string
  color: string
  drawAt: string
}

export type PlayHistoryRow = {
  periodNo: string
  result: string
  stake: string
  payout: string
  status: 'WON' | 'LOST' | 'PENDING'
  settledAt: string
}

export type PlayBetOption = {
  key: string
  label: string
  hint?: string
  accent: string
}

export type PlayBetGroup = {
  title: string
  description: string
  mode: 'chips' | 'grid'
  options: PlayBetOption[]
}

export type PlayVariant = {
  code: string
  label: string
  durationLabel: string
  periodNo: string
  serverTime: string
  openAt: string
  closeAt: string
  drawAt: string
  countdownSeconds: number
  recentResults: PlayResult[]
  history: PlayHistoryRow[]
  myHistory: PlayHistoryRow[]
  betGroups: PlayBetGroup[]
  note: string
}

export type PlayRoom = {
  code: string
  title: string
  subtitle: string
  accent: string
  symbol: string
  category: string
  minBet: string
  payout: string
  jackpot: string
  onlinePlayers: number
  status: 'OPEN' | 'COMING_SOON'
  featured?: boolean
  variants: PlayVariant[]
  rules: string[]
}

export const playCategories = ['Tất cả', 'Win Go', 'K3', 'Lô tô', 'Sắp mở']

function nowIso(): string {
  return new Date().toISOString()
}

function addSeconds(seconds: number): string {
  return new Date(Date.now() + seconds * 1000).toISOString()
}

function buildHistory(prefix: string, rows: Array<[string, string, string, string, 'WON' | 'LOST' | 'PENDING']>): PlayHistoryRow[] {
  return rows.map(([periodNo, result, stake, payout, status], index) => ({
    periodNo: `${prefix}-${periodNo}`,
    result,
    stake,
    payout,
    status,
    settledAt: addSeconds(-index * 180),
  }))
}

function buildWingoResults(base: string): PlayResult[] {
  return [
    { periodNo: `${base}-342`, result: '7', bigSmall: 'Lớn', color: 'Xanh', drawAt: addSeconds(-20) },
    { periodNo: `${base}-341`, result: '2', bigSmall: 'Nhỏ', color: 'Đỏ', drawAt: addSeconds(-50) },
    { periodNo: `${base}-340`, result: '0', bigSmall: 'Nhỏ', color: 'Tím', drawAt: addSeconds(-80) },
    { periodNo: `${base}-339`, result: '5', bigSmall: 'Lớn', color: 'Tím', drawAt: addSeconds(-110) },
    { periodNo: `${base}-338`, result: '9', bigSmall: 'Lớn', color: 'Xanh', drawAt: addSeconds(-140) },
  ]
}

function buildK3Results(base: string): PlayResult[] {
  return [
    { periodNo: `${base}-442`, result: '4-4-4', bigSmall: 'Bộ ba', color: 'Đỏ', drawAt: addSeconds(-20) },
    { periodNo: `${base}-441`, result: '1-2-6', bigSmall: 'Tổng 9', color: 'Xanh', drawAt: addSeconds(-50) },
    { periodNo: `${base}-440`, result: '3-3-1', bigSmall: 'Tổng 7', color: 'Tím', drawAt: addSeconds(-80) },
    { periodNo: `${base}-439`, result: '6-5-4', bigSmall: 'Tổng 15', color: 'Xanh', drawAt: addSeconds(-110) },
    { periodNo: `${base}-438`, result: '2-2-8', bigSmall: 'Tổng 12', color: 'Đỏ', drawAt: addSeconds(-140) },
  ]
}

function buildLotteryResults(base: string): PlayResult[] {
  return [
    { periodNo: `${base}-542`, result: '12345', bigSmall: 'Tổng 15', color: 'Xanh', drawAt: addSeconds(-20) },
    { periodNo: `${base}-541`, result: '90876', bigSmall: 'Tổng 30', color: 'Đỏ', drawAt: addSeconds(-50) },
    { periodNo: `${base}-540`, result: '55678', bigSmall: 'Tổng 31', color: 'Tím', drawAt: addSeconds(-80) },
    { periodNo: `${base}-539`, result: '11223', bigSmall: 'Tổng 9', color: 'Xanh', drawAt: addSeconds(-110) },
    { periodNo: `${base}-538`, result: '67890', bigSmall: 'Tổng 30', color: 'Đỏ', drawAt: addSeconds(-140) },
  ]
}

function buildWingoVariant(code: string, label: string, durationLabel: string, countdownSeconds: number): PlayVariant {
  const base = `WG-${code.toUpperCase()}-${Date.now().toString().slice(-5)}`
  return {
    code,
    label,
    durationLabel,
    periodNo: `${base}-343`,
    serverTime: nowIso(),
    openAt: addSeconds(-15),
    closeAt: addSeconds(Math.max(countdownSeconds - 5, 5)),
    drawAt: addSeconds(countdownSeconds),
    countdownSeconds,
    recentResults: buildWingoResults(base),
    history: buildHistory('WG', [
      ['342', '7 / Xanh / Lớn', '50.000đ', '98.000đ', 'WON'],
      ['341', '2 / Đỏ / Nhỏ', '20.000đ', '0đ', 'LOST'],
      ['340', '0 / Tím / Nhỏ', '10.000đ', '45.000đ', 'WON'],
      ['339', '5 / Tím / Lớn', '10.000đ', '45.000đ', 'WON'],
    ]),
    myHistory: buildHistory('WG-ME', [
      ['342', 'Xanh', '100.000đ', '196.000đ', 'WON'],
      ['341', 'Đỏ', '20.000đ', '0đ', 'LOST'],
      ['340', 'Tím', '10.000đ', '45.000đ', 'WON'],
    ]),
    betGroups: [
      {
        title: 'Màu sắc',
        description: 'Chọn Xanh, Đỏ hoặc Tím theo rule Win Go.',
        mode: 'chips',
        options: [
          { key: 'green', label: 'Xanh', accent: '#24b561' },
          { key: 'red', label: 'Đỏ', accent: '#e64545' },
          { key: 'violet', label: 'Tím', accent: '#8b5cf6' },
        ],
      },
      {
        title: 'Chọn số',
        description: 'Chọn 1 số từ 0 đến 9 cho kỳ hiện tại.',
        mode: 'grid',
        options: Array.from({ length: 10 }, (_, number) => ({
          key: `number_${number}`,
          label: String(number),
          accent: number % 3 === 0 ? '#24b561' : number % 3 === 1 ? '#e64545' : '#8b5cf6',
        })),
      },
      {
        title: 'Lớn / Nhỏ',
        description: 'Cửa cược đơn giản cho Win Go.',
        mode: 'chips',
        options: [
          { key: 'big', label: 'LỚN', accent: '#f6c32d' },
          { key: 'small', label: 'NHỎ', accent: '#24b561' },
        ],
      },
    ],
    note: 'Win Go có nhịp nhanh, ưu tiên đặt lệnh sớm trước khi khóa kỳ.',
  }
}

function buildK3Variant(code: string, label: string, durationLabel: string, countdownSeconds: number): PlayVariant {
  const base = `K3-${code.toUpperCase()}-${Date.now().toString().slice(-5)}`
  return {
    code,
    label,
    durationLabel,
    periodNo: `${base}-443`,
    serverTime: nowIso(),
    openAt: addSeconds(-20),
    closeAt: addSeconds(Math.max(countdownSeconds - 8, 8)),
    drawAt: addSeconds(countdownSeconds),
    countdownSeconds,
    recentResults: buildK3Results(base),
    history: buildHistory('K3', [
      ['442', '4-4-4', '50.000đ', '450.000đ', 'WON'],
      ['441', '1-2-6', '20.000đ', '0đ', 'LOST'],
      ['440', '3-3-1', '50.000đ', '0đ', 'LOST'],
      ['439', '6-5-4', '10.000đ', '52.000đ', 'WON'],
    ]),
    myHistory: buildHistory('K3-ME', [
      ['442', 'Bộ ba 4', '30.000đ', '270.000đ', 'WON'],
      ['441', 'Tổng 9', '20.000đ', '0đ', 'LOST'],
      ['440', 'Lớn', '10.000đ', '19.000đ', 'WON'],
    ]),
    betGroups: [
      {
        title: 'Tổng điểm',
        description: 'Chọn tổng 3 đến 18 theo bộ 3 xúc xắc.',
        mode: 'chips',
        options: Array.from({ length: 16 }, (_, index) => {
          const total = index + 3
          return { key: `sum_${total}`, label: String(total), accent: total >= 11 ? '#e64545' : '#24b561' }
        }),
      },
      {
        title: 'Chẵn / Lẻ',
        description: 'Bám theo tổng kết quả của K3.',
        mode: 'chips',
        options: [
          { key: 'odd', label: 'Lẻ', accent: '#8b5cf6' },
          { key: 'even', label: 'Chẵn', accent: '#f6c32d' },
          { key: 'big', label: 'Lớn', accent: '#f6c32d' },
          { key: 'small', label: 'Nhỏ', accent: '#24b561' },
        ],
      },
      {
        title: 'Bộ ba',
        description: 'Cược bộ ba toàn bộ hoặc bộ ba cụ thể.',
        mode: 'chips',
        options: [
          { key: 'triple_any', label: 'Any Triple', accent: '#f6c32d' },
          { key: 'triple_111', label: '111', accent: '#e64545' },
          { key: 'triple_222', label: '222', accent: '#e64545' },
          { key: 'triple_333', label: '333', accent: '#e64545' },
          { key: 'triple_444', label: '444', accent: '#e64545' },
          { key: 'triple_555', label: '555', accent: '#e64545' },
          { key: 'triple_666', label: '666', accent: '#e64545' },
        ],
      },
    ],
    note: 'K3 cho phép nhiều cửa theo tổng và bộ ba, nhưng UI phải giữ cho chọn lệnh thật nhanh.',
  }
}

function buildLotteryVariant(code: string, label: string, durationLabel: string, countdownSeconds: number): PlayVariant {
  const base = `5D-${code.toUpperCase()}-${Date.now().toString().slice(-5)}`
  return {
    code,
    label,
    durationLabel,
    periodNo: `${base}-543`,
    serverTime: nowIso(),
    openAt: addSeconds(-20),
    closeAt: addSeconds(Math.max(countdownSeconds - 10, 10)),
    drawAt: addSeconds(countdownSeconds),
    countdownSeconds,
    recentResults: buildLotteryResults(base),
    history: buildHistory('5D', [
      ['542', '12345', '50.000đ', '490.000đ', 'WON'],
      ['541', '90876', '20.000đ', '0đ', 'LOST'],
      ['540', '55678', '100.000đ', '0đ', 'LOST'],
      ['539', '11223', '50.000đ', '245.000đ', 'WON'],
    ]),
    myHistory: buildHistory('5D-ME', [
      ['542', 'A=1, B=2, C=3', '100.000đ', '490.000đ', 'WON'],
      ['541', 'Tổng 30', '50.000đ', '0đ', 'LOST'],
      ['540', 'Lớn / Chẵn', '20.000đ', '0đ', 'LOST'],
    ]),
    betGroups: [
      {
        title: 'Vị trí A - E',
        description: 'Chọn 1 chữ số cho từng vị trí.',
        mode: 'grid',
        options: Array.from({ length: 10 }, (_, number) => ({
          key: `digit_${number}`,
          label: String(number),
          accent: number >= 5 ? '#f6c32d' : '#24b561',
        })),
      },
      {
        title: 'Tổng hợp',
        description: 'Đặt theo tổng lớn/nhỏ hoặc chẵn/lẻ.',
        mode: 'chips',
        options: [
          { key: 'big', label: 'LỚN', accent: '#f6c32d' },
          { key: 'small', label: 'NHỎ', accent: '#24b561' },
          { key: 'odd', label: 'LẺ', accent: '#8b5cf6' },
          { key: 'even', label: 'CHẴN', accent: '#e64545' },
        ],
      },
      {
        title: 'Số đuôi',
        description: 'Cược theo số cuối hoặc tổng chữ số.',
        mode: 'chips',
        options: [
          { key: 'last_0', label: 'Đuôi 0', accent: '#e64545' },
          { key: 'last_5', label: 'Đuôi 5', accent: '#e64545' },
          { key: 'sum_15', label: 'Tổng 15', accent: '#f6c32d' },
          { key: 'sum_30', label: 'Tổng 30', accent: '#f6c32d' },
        ],
      },
    ],
    note: '5D ưu tiên chọn vị trí và tổng hợp, các cửa phức tạp sẽ mở dần theo cấu hình vận hành.',
  }
}

export const playRooms: PlayRoom[] = [
  {
    code: 'wingo',
    title: 'Win Go',
    subtitle: 'Phòng 1 room duy nhất cho từng nhịp 30 giây, 1 phút, 3 phút và 5 phút.',
    accent: '#ff6d66',
    symbol: 'rocket_launch',
    category: 'Win Go',
    minBet: '10.000đ',
    payout: '98.6%',
    jackpot: '125.000.000đ',
    onlinePlayers: 1834,
    status: 'OPEN',
    featured: true,
    variants: [
      buildWingoVariant('30s', 'Win Go 30 giây', '30 giây', 28),
      buildWingoVariant('1m', 'Win Go 1 phút', '1 phút', 58),
      buildWingoVariant('3m', 'Win Go 3 phút', '3 phút', 178),
      buildWingoVariant('5m', 'Win Go 5 phút', '5 phút', 298),
    ],
    rules: [
      'Không đặt 2 bên đối lập trong cùng kỳ.',
      'Không vượt quá 8 số trong một kỳ.',
      'Stake cuối là giá trị FE gửi lên backend.',
    ],
  },
  {
    code: 'k3',
    title: 'K3',
    subtitle: 'Xúc xắc 3 viên, phù hợp các cửa tổng, bộ ba và chẵn/lẻ.',
    accent: '#e64545',
    symbol: 'casino',
    category: 'K3',
    minBet: '20.000đ',
    payout: '97.9%',
    jackpot: '80.000.000đ',
    onlinePlayers: 1021,
    status: 'OPEN',
    featured: true,
    variants: [
      buildK3Variant('1m', 'K3 1 phút', '1 phút', 58),
      buildK3Variant('3m', 'K3 3 phút', '3 phút', 178),
      buildK3Variant('5m', 'K3 5 phút', '5 phút', 298),
      buildK3Variant('10m', 'K3 10 phút', '10 phút', 598),
    ],
    rules: [
      'Cửa tổng, chẵn/lẻ và lớn/nhỏ phải theo đúng result payload.',
      'Bộ ba, đôi, liên tiếp sẽ được settlement theo rule từng variant.',
      'Request không hợp lệ phải bị từ chối trước khi đặt lệnh.',
    ],
  },
  {
    code: 'lottery',
    title: '5D Lô tô',
    subtitle: 'Chọn vị trí, tổng hợp và số đuôi theo vòng quay 5 số.',
    accent: '#f6c32d',
    symbol: 'looks_5',
    category: 'Lô tô',
    minBet: '50.000đ',
    payout: '96.8%',
    jackpot: '250.000.000đ',
    onlinePlayers: 856,
    status: 'OPEN',
    featured: true,
    variants: [
      buildLotteryVariant('1m', '5D 1 phút', '1 phút', 58),
      buildLotteryVariant('3m', '5D 3 phút', '3 phút', 178),
      buildLotteryVariant('5m', '5D 5 phút', '5 phút', 298),
      buildLotteryVariant('10m', '5D 10 phút', '10 phút', 598),
    ],
    rules: [
      'Không đặt cược đối lập trong cùng kỳ.',
      'Chọn vị trí A - E hoặc tổng hợp theo rule vận hành.',
      'Kết quả phải được normalize về digits + sum + last digit.',
    ],
  },
  {
    code: 'trx_win',
    title: 'Trx Win',
    subtitle: 'Màn mở rộng hệ sinh thái, sẽ kích hoạt sau khi có cấu hình riêng.',
    accent: '#24b561',
    symbol: 'currency_bitcoin',
    category: 'Sắp mở',
    minBet: '--',
    payout: '--',
    jackpot: '--',
    onlinePlayers: 0,
    status: 'COMING_SOON',
    variants: [],
    rules: ['Chưa mở room vận hành.'],
  },
]

export function getPlayRoom(code: string) {
  return playRooms.find((room) => room.code === code)
}

export function getPlayVariant(roomCode: string, variantCode: string) {
  return getPlayRoom(roomCode)?.variants.find((variant) => variant.code === variantCode)
}
