<script setup lang="ts">
import { computed, ref } from 'vue'

import { formatViMoney } from '@/shared/lib/money'

interface BetStat {
  option_key: string
  option_type: number
  player_count: number
  total_stake: string
}

interface RoomStats {
  code: string
  game_type: number
  period: {
    id: number
    period_no: string
    period_index: number
    draw_at: string
    bet_lock_at: string
    status: number
    manual_result: string | null
  } | null
  bet_stats: BetStat[]
}

type SubmitPayload = {
  roomCode: string
  periodId: number
  result: string
  bigSmall: string
  color: string
  payload: Record<string, any>
  title: string
}

const props = defineProps<{
  room: RoomStats
  nowMs: number
  isSubmitting: boolean
  pulsing?: boolean
  manualStateLabel?: string
}>()

const emit = defineEmits<{
  (e: 'request-submit', payload: SubmitPayload): void
}>()

const k3Dice = ref<[number, number, number]>([1, 1, 1])
const lotteryDigits = ref<[number, number, number, number, number]>([0, 0, 0, 0, 0])

const stats = computed(() => props.room.bet_stats ?? [])

type AggregatedOptionStat = {
  key: string
  label: string
  stake: number
  players: number
}

function parseStake(value: string): number {
  const num = Number.parseFloat(value ?? '0')
  return Number.isFinite(num) ? num : 0
}

function stakeByKey(key: string): number {
  const lower = key.toLowerCase()
  return stats.value
    .filter((item) => item.option_key.toLowerCase() === lower)
    .reduce((sum, item) => sum + parseStake(item.total_stake), 0)
}

function optionStatByKey(key: string, label = formatBetOptionLabel(key)): AggregatedOptionStat {
  const normalizedKey = key.toLowerCase()
  const matched = stats.value.filter((item) => item.option_key.toLowerCase() === normalizedKey)
  return {
    key: normalizedKey,
    label,
    stake: matched.reduce((sum, item) => sum + parseStake(item.total_stake), 0),
    players: matched.reduce((sum, item) => sum + Number(item.player_count || 0), 0),
  }
}

function buildOptionStats(keys: string[]): AggregatedOptionStat[] {
  return keys.map((key) => optionStatByKey(key))
}

function formatBetOptionLabel(rawKey: string): string {
  const key = String(rawKey || '').trim().toLowerCase()
  if (!key) return '---'

  if (key.startsWith('number_')) return `Ball ${key.replace('number_', '')}`
  if (key === 'big') return 'Lớn'
  if (key === 'small') return 'Nhỏ'
  if (key === 'odd') return 'Lẻ'
  if (key === 'even') return 'Chẵn'
  if (key === 'green') return 'Xanh'
  if (key === 'red') return 'Đỏ'
  if (key === 'violet') return 'Tím'
  if (key === 'green_violet') return 'Xanh / Tím'
  if (key === 'red_violet') return 'Đỏ / Tím'

  if (key.startsWith('pair_')) return `Một đôi ${key.replace('pair_', '')}`
  if (key.startsWith('triple_')) return `Bộ ba ${key.replace('triple_', '')}`
  if (key === 'serial_any') return '3 số liên tiếp'
  if (key.startsWith('diff_')) return `3 số khác nhau chứa ${key.replace('diff_', '')}`
  if (/^sum_\d+$/.test(key)) return `Tổng ${key.replace('sum_', '')}`

  const lotteryPositionDigitMatch = key.match(/^pos_([a-e])_(\d)$/)
  if (lotteryPositionDigitMatch) {
    const position = lotteryPositionDigitMatch[1]?.toUpperCase() ?? ''
    const digit = lotteryPositionDigitMatch[2] ?? ''
    return `Vị trí ${position} = ${digit}`
  }

  const lotteryPositionPropertyMatch = key.match(/^pos_([a-e])_(big|small|odd|even)$/)
  if (lotteryPositionPropertyMatch) {
    const position = lotteryPositionPropertyMatch[1]?.toUpperCase() ?? ''
    const property = lotteryPositionPropertyMatch[2] ?? ''
    return `Vị trí ${position} • ${formatBetOptionLabel(property)}`
  }

  const lotterySumPropertyMatch = key.match(/^sum_(big|small|odd|even)$/)
  if (lotterySumPropertyMatch) {
    const property = lotterySumPropertyMatch[1] ?? ''
    return `SUM • ${formatBetOptionLabel(property)}`
  }

  if (key.startsWith('last_')) return `Đuôi ${key.replace('last_', '')}`
  if (key.startsWith('pick5_')) return `5 số ${key.replace('pick5_', '')}`

  return rawKey.replaceAll('_', ' ')
}

function compactOptionLabel(rawLabel: string): string {
  return rawLabel
    .replace(/^Vị trí [A-E] • /, '')
    .replace(/^SUM • /, '')
    .trim()
}

const totalStake = computed(() =>
  stats.value.reduce((sum, item) => sum + parseStake(item.total_stake), 0),
)

const optionStakeRows = computed<AggregatedOptionStat[]>(() => {
  const grouped = new Map<string, AggregatedOptionStat>()

  for (const item of stats.value) {
    const key = String(item.option_key || '').trim().toLowerCase()
    if (!key) continue

    const current = grouped.get(key) ?? {
      key,
      label: formatBetOptionLabel(key),
      stake: 0,
      players: 0,
    }

    current.stake += parseStake(item.total_stake)
    current.players += Number(item.player_count || 0)
    grouped.set(key, current)
  }

  return Array.from(grouped.values()).sort((a, b) => {
    if (b.stake !== a.stake) return b.stake - a.stake
    return a.label.localeCompare(b.label, 'vi')
  })
})

function secondsLeft(): number {
  const drawAt = props.room.period?.draw_at
  if (!drawAt) return 0
  const cleanDrawAt = drawAt.substring(0, 19).replace(' ', 'T')
  const drawMs = new Date(cleanDrawAt).getTime()
  if (!Number.isFinite(drawMs)) return 0
  return Math.max(0, Math.floor((drawMs - props.nowMs) / 1000))
}

const leftSeconds = computed(() => secondsLeft())
const countdownLabel = computed(() => {
  const sec = leftSeconds.value
  const mm = Math.floor(sec / 60)
  const ss = sec % 60
  return `${String(mm).padStart(2, '0')}:${String(ss).padStart(2, '0')}`
})

const isLocked = computed(() => {
  const period = props.room.period
  if (!period) return true
  if (period.status >= 4) return true
  return leftSeconds.value <= 0
})

const lockMessage = computed(() => {
  if (!props.room.period) return 'Chờ kỳ mới'
  if (props.room.period.status >= 4) return 'Đã khóa can thiệp'
  if (leftSeconds.value <= 0) return 'Đã hết thời gian can thiệp'
  if (leftSeconds.value <= 3) return `Can thiệp còn ${leftSeconds.value}s`
  return 'Đang mở can thiệp'
})

const manualResult = computed(() => {
  if (!props.room.period?.manual_result) return null
  try {
    return JSON.parse(props.room.period.manual_result) as Record<string, any>
  } catch {
    return null
  }
})

const manualResultSummary = computed(() => {
  if (props.manualStateLabel) return props.manualStateLabel
  const result = String(manualResult.value?.result || '').trim()
  if (!result) return ''
  return `KQ can thiệp: ${result}`
})

const winGoNumbers = computed(() =>
  Array.from({ length: 10 }, (_, num) => ({
    number: num,
    stake: stakeByKey(`number_${num}`),
    players: stats.value
      .filter((item) => item.option_key.toLowerCase() === `number_${num}`)
      .reduce((sum, item) => sum + Number(item.player_count || 0), 0),
  })),
)

const maxWinGoNumberStake = computed(() =>
  Math.max(1, ...winGoNumbers.value.map((item) => item.stake)),
)

function oddsForBetOption(optionKey: string): number {
  const key = optionKey.toLowerCase()
  if (key.startsWith('number_')) return 9
  if (key.startsWith('digit_')) return 9
  if (key.startsWith('last_')) return 9
  if (key === 'violet') return 4.5
  if (key === 'green' || key === 'red' || key === 'big' || key === 'small' || key === 'odd' || key === 'even') {
    return 2
  }
  if (key.startsWith('pair_')) return 13.83
  if (key.startsWith('triple_')) return 207.36
  if (key === 'serial_any') return 8.64
  if (key.startsWith('diff_')) return 34.56
  if (key.startsWith('sum_')) {
    switch (key) {
      case 'sum_3':
      case 'sum_18':
        return 207.36
      case 'sum_4':
      case 'sum_17':
        return 69.12
      case 'sum_5':
      case 'sum_16':
        return 34.56
      case 'sum_6':
      case 'sum_15':
      case 'sum_30':
        return 20.74
      case 'sum_7':
      case 'sum_14':
        return 13.83
      case 'sum_8':
      case 'sum_13':
        return 9.88
      case 'sum_9':
      case 'sum_12':
        return 8.3
      case 'sum_10':
      case 'sum_11':
        return 7.68
      default:
        return 2
    }
  }
  return 2
}

const wingoPrimaryCards = computed(() => [
  optionStatByKey('big', 'Lớn'),
  optionStatByKey('small', 'Nhỏ'),
  optionStatByKey('odd', 'Lẻ'),
  optionStatByKey('even', 'Chẵn'),
])

const colorCards = computed(() => [
  optionStatByKey('green', 'Xanh'),
  optionStatByKey('red', 'Đỏ'),
  optionStatByKey('violet', 'Tím'),
])

function calculateWinGoPL(outcome: number): number {
  const bigSmall = outcome >= 5 ? 'big' : 'small'
  const oddEven = outcome % 2 === 0 ? 'even' : 'odd'
  const colorTags: string[] = []
  if (outcome === 0) colorTags.push('red', 'violet')
  else if (outcome === 5) colorTags.push('green', 'violet')
  else colorTags.push(outcome % 2 === 0 ? 'red' : 'green')

  const winTags = new Set<string>([`number_${outcome}`, bigSmall, oddEven, ...colorTags])
  let payout = 0
  const taxRate = 0.02

  for (const stat of stats.value) {
    if (winTags.has(stat.option_key.toLowerCase())) {
      payout += parseStake(stat.total_stake) * (oddsForBetOption(stat.option_key) - taxRate)
    }
  }

  return totalStake.value - payout
}

const safestWingo = computed(() => {
  let best = { number: 0, pl: Number.NEGATIVE_INFINITY }
  for (let n = 0; n <= 9; n += 1) {
    const pl = calculateWinGoPL(n)
    if (pl > best.pl) best = { number: n, pl }
  }
  return best
})

function requestWinGo(number: number): void {
  const periodId = props.room.period?.id
  if (!periodId || isLocked.value || props.isSubmitting) return
  const bigSmall = number >= 5 ? 'big' : 'small'
  let color = ''
  const tags: string[] = [`number_${number}`, bigSmall, number % 2 === 0 ? 'even' : 'odd']
  if (number === 0) {
    color = 'red_violet'
    tags.push('red', 'violet')
  } else if (number === 5) {
    color = 'green_violet'
    tags.push('green', 'violet')
  } else {
    color = number % 2 === 0 ? 'red' : 'green'
    tags.push(color)
  }

  emit('request-submit', {
    roomCode: props.room.code,
    periodId,
    result: String(number),
    bigSmall,
    color,
    payload: {
      game_type: 'wingo',
      number,
      result: String(number),
      big_small: bigSmall,
      color,
      tags,
      generated_at: new Date().toISOString(),
    },
    title: `WinGo #${number}`,
  })
}

const k3Sums = computed(() =>
  Array.from({ length: 16 }, (_, idx) => {
    const sum = idx + 3
    return { sum, stake: stakeByKey(`sum_${sum}`) }
  }),
)

const maxK3SumStake = computed(() => Math.max(1, ...k3Sums.value.map((item) => item.stake)))

function sumDice(dice: number[]): number {
  return dice.reduce((sum, d) => sum + d, 0)
}

function buildK3Outcome(dice: number[]) {
  const normalizedDice: [number, number, number] = [
    Number(dice[0] ?? 1),
    Number(dice[1] ?? 1),
    Number(dice[2] ?? 1),
  ]
  const sum = sumDice(normalizedDice)
  const bigSmall = sum >= 11 ? 'big' : 'small'
  const oddEven = sum % 2 === 0 ? 'even' : 'odd'
  const counts = new Map<number, number>()
  for (const value of normalizedDice) {
    counts.set(value, (counts.get(value) ?? 0) + 1)
  }

  const sortedUniqueValues = Array.from(counts.keys()).sort((a, b) => a - b)
  const tags: string[] = [`sum_${sum}`, bigSmall, oddEven]
  const isTriple = counts.size === 1

  if (isTriple) {
    tags.push('triple_any', `triple_${normalizedDice[0]}`)
  } else if (counts.size === 2) {
    const pairValue = sortedUniqueValues.find((value) => counts.get(value) === 2)
    if (pairValue != null) {
      tags.push(`pair_${pairValue}`)
    }
  } else {
    if (
      sortedUniqueValues[0] != null &&
      sortedUniqueValues[1] != null &&
      sortedUniqueValues[2] != null &&
      sortedUniqueValues[0] + 1 === sortedUniqueValues[1] &&
      sortedUniqueValues[1] + 1 === sortedUniqueValues[2]
    ) {
      tags.push('serial_any')
    }
    for (const value of sortedUniqueValues) {
      tags.push(`diff_${value}`)
    }
  }

  return {
    dice: normalizedDice,
    sum,
    result: normalizedDice.join('-'),
    bigSmall,
    oddEven,
    isTriple,
    tags: Array.from(new Set(tags)),
  }
}

function buildLotteryOutcome(digits: number[]) {
  const normalizedDigits: [number, number, number, number, number] = [
    Number(digits[0] ?? 0),
    Number(digits[1] ?? 0),
    Number(digits[2] ?? 0),
    Number(digits[3] ?? 0),
    Number(digits[4] ?? 0),
  ]
  const sum = normalizedDigits.reduce((total, digit) => total + digit, 0)
  const last = normalizedDigits[4] ?? 0
  const bigSmall = last >= 5 ? 'big' : 'small'
  const oddEven = last % 2 === 0 ? 'even' : 'odd'
  const sumBigSmall = sum >= 23 ? 'big' : 'small'
  const sumOddEven = sum % 2 === 0 ? 'even' : 'odd'
  const positions: Record<string, { digit: number; big_small: string; odd_even: string }> = {}
  const tags: string[] = [
    `pick5_${normalizedDigits.join('')}`,
    `sum_${sum}`,
    `last_${last}`,
    bigSmall,
    oddEven,
    `sum_${sumBigSmall}`,
    `sum_${sumOddEven}`,
  ]

  normalizedDigits.forEach((digit, index) => {
    const position = String.fromCharCode(97 + index)
    const upper = position.toUpperCase()
    const positionBigSmall = digit >= 5 ? 'big' : 'small'
    const positionOddEven = digit % 2 === 0 ? 'even' : 'odd'
    positions[upper] = {
      digit,
      big_small: positionBigSmall,
      odd_even: positionOddEven,
    }
    tags.push(`pos_${position}_${digit}`, `pos_${position}_${positionBigSmall}`, `pos_${position}_${positionOddEven}`)
  })

  return {
    digits: normalizedDigits,
    positions,
    sum,
    sumBigSmall,
    sumOddEven,
    last,
    result: normalizedDigits.join(''),
    bigSmall,
    oddEven,
    tags: Array.from(new Set(tags)),
  }
}

function calculateK3PL(sum: number): number {
  const bigSmall = sum >= 11 ? 'big' : 'small'
  const oddEven = sum % 2 === 0 ? 'even' : 'odd'
  const winTags = new Set<string>([`sum_${sum}`, bigSmall, oddEven])
  let payout = 0
  const taxRate = 0.02
  for (const stat of stats.value) {
    if (winTags.has(stat.option_key.toLowerCase())) {
      payout += parseStake(stat.total_stake) * (oddsForBetOption(stat.option_key) - taxRate)
    }
  }
  return totalStake.value - payout
}

function diceForSum(target: number): [number, number, number] {
  for (let a = 1; a <= 6; a += 1) {
    for (let b = 1; b <= 6; b += 1) {
      for (let c = 1; c <= 6; c += 1) {
        if (a + b + c === target) return [a, b, c]
      }
    }
  }
  return [1, 1, 1]
}

const safestK3 = computed(() => {
  let best = { sum: 3, pl: Number.NEGATIVE_INFINITY }
  for (let s = 3; s <= 18; s += 1) {
    const pl = calculateK3PL(s)
    if (pl > best.pl) best = { sum: s, pl }
  }
  return best
})

function applySafestK3(): void {
  k3Dice.value = diceForSum(safestK3.value.sum)
}

function requestK3(): void {
  const periodId = props.room.period?.id
  if (!periodId || isLocked.value || props.isSubmitting) return
  const outcome = buildK3Outcome(k3Dice.value)
  emit('request-submit', {
    roomCode: props.room.code,
    periodId,
    result: outcome.result,
    bigSmall: outcome.bigSmall,
    color: '-',
    payload: {
      game_type: 'k3',
      dice: outcome.dice,
      sum: outcome.sum,
      result: outcome.result,
      big_small: outcome.bigSmall,
      odd_even: outcome.oddEven,
      is_triple: outcome.isTriple,
      tags: outcome.tags,
      generated_at: new Date().toISOString(),
    },
    title: `K3 ${outcome.result}`,
  })
}

const k3PrimaryCards = computed(() => buildOptionStats(['big', 'small', 'odd', 'even']))
const k3PairCards = computed(() => buildOptionStats(['pair_1', 'pair_2', 'pair_3', 'pair_4', 'pair_5', 'pair_6']))
const k3TripleCards = computed(() => buildOptionStats(['triple_1', 'triple_2', 'triple_3', 'triple_4', 'triple_5', 'triple_6']))
const k3DiffCards = computed(() => buildOptionStats(['serial_any', 'diff_1', 'diff_2', 'diff_3', 'diff_4', 'diff_5', 'diff_6']))

type MatrixCell = { key: string; stake: number }

const lotteryMatrix = computed<MatrixCell[][]>(() => {
  const rows: MatrixCell[][] = ['a', 'b', 'c', 'd', 'e'].map((row) =>
    Array.from({ length: 10 }, (_, digit) => ({
      key: `${row}_${digit}`,
      stake: 0,
    })),
  )

  for (const stat of stats.value) {
    const key = stat.option_key.toLowerCase()
    const m =
      key.match(/(?:position_|pos_|digit_)?([abcde])[_-]?(\d)$/) ||
      key.match(/^([abcde])(\d)$/)
    if (!m) continue
    const rowChar = m[1]
    const digitStr = m[2]
    if (!rowChar || !digitStr) continue

    const rowIdx = 'abcde'.indexOf(rowChar)
    const digit = Number.parseInt(digitStr, 10)
    if (rowIdx < 0 || digit < 0 || digit > 9) continue

    const row = rows[rowIdx]
    if (row && row[digit]) {
      row[digit].stake += parseStake(stat.total_stake)
    }
  }

  return rows
})

const lotteryRowMax = computed(() =>
  lotteryMatrix.value.map((row) => Math.max(1, ...row.map((item) => item.stake))),
)

const lotteryPositionPropertyRows = computed(() =>
  ['a', 'b', 'c', 'd', 'e'].map((position) => ({
    position: position.toUpperCase(),
    items: buildOptionStats([
      `pos_${position}_big`,
      `pos_${position}_small`,
      `pos_${position}_odd`,
      `pos_${position}_even`,
    ]),
  })),
)

const lotterySumPropertyCards = computed(() => buildOptionStats(['sum_big', 'sum_small', 'sum_odd', 'sum_even']))

function calculateLotteryPL(digits: number[]): number {
  const result = digits.join('')
  const sum = digits.reduce((a, b) => a + b, 0)
  const last = digits[4] ?? 0
  const bigSmall = last >= 5 ? 'big' : 'small'
  const oddEven = last % 2 === 0 ? 'even' : 'odd'
  const winTags = new Set<string>([`pick5_${result}`, `sum_${sum}`, `last_${last}`, bigSmall, oddEven])

  let payout = 0
  const taxRate = 0.02
  for (const stat of stats.value) {
    if (winTags.has(stat.option_key.toLowerCase())) {
      payout += parseStake(stat.total_stake) * (oddsForBetOption(stat.option_key) - taxRate)
    }
  }
  return totalStake.value - payout
}

function requestLottery(): void {
  const periodId = props.room.period?.id
  if (!periodId || isLocked.value || props.isSubmitting) return
  const outcome = buildLotteryOutcome(lotteryDigits.value)
  emit('request-submit', {
    roomCode: props.room.code,
    periodId,
    result: outcome.result,
    bigSmall: outcome.bigSmall,
    color: '-',
    payload: {
      game_type: 'lottery',
      digits: outcome.digits,
      positions: outcome.positions,
      sum: outcome.sum,
      sum_big_small: outcome.sumBigSmall,
      sum_odd_even: outcome.sumOddEven,
      last_digit: outcome.last,
      result: outcome.result,
      big_small: outcome.bigSmall,
      odd_even: outcome.oddEven,
      tags: outcome.tags,
      generated_at: new Date().toISOString(),
    },
    title: `5D ${outcome.result}`,
  })
}

const previewPL = computed(() => {
  if (props.room.game_type === 1) return calculateWinGoPL(safestWingo.value.number)
  if (props.room.game_type === 2) return calculateK3PL(sumDice(k3Dice.value))
  return calculateLotteryPL(lotteryDigits.value)
})

const riskClass = computed(() => (previewPL.value < 0 ? 'risk-alert' : 'risk-safe'))

function heatHeight(stake: number, maxStake: number): string {
  const ratioValue = Math.max(0.08, Math.min(1, stake / maxStake))
  return `${Math.round(ratioValue * 100)}%`
}
</script>

<template>
  <article class="glass-card control-room-card" :class="[riskClass, pulsing ? 'card-pulse' : '']">
    <header class="room-head">
      <div>
        <div class="room-title-row">
          <h3 class="room-title">{{ room.code.replace('_', ' ').toUpperCase() }}</h3>
          <span class="room-chip" :class="isLocked ? 'room-chip--lock' : 'room-chip--open'">{{ lockMessage }}</span>
        </div>
        <p class="room-period">Kỳ: {{ room.period?.period_index || room.period?.period_no || '---' }}</p>
        <p v-if="manualResultSummary" class="room-manual-state">{{ manualResultSummary }}</p>
      </div>
      <div class="room-timer">
        <p class="room-timer__label">Time Left</p>
        <strong class="room-timer__value">{{ countdownLabel }}</strong>
      </div>
    </header>

    <div class="room-body">
      <section v-if="room.game_type === 1" class="board board-wingo">
        <div class="metric-grid metric-grid--four">
          <article v-for="card in wingoPrimaryCards" :key="card.key" class="metric-card">
            <p class="metric-card__label">{{ card.label }}</p>
            <strong class="metric-card__value">{{ formatViMoney(card.stake.toString()) }}</strong>
            <span class="metric-card__meta">{{ card.players.toLocaleString('vi-VN') }} lượt</span>
          </article>
        </div>

        <div class="color-cards">
          <div v-for="card in colorCards" :key="card.key" class="color-card" :class="`color-card--${card.key}`">
            <p>{{ card.label }}</p>
            <strong>{{ formatViMoney(card.stake.toString()) }}</strong>
            <span>{{ card.players.toLocaleString('vi-VN') }} lượt</span>
          </div>
        </div>

        <div class="numbers-grid">
          <button
            v-for="item in winGoNumbers"
            :key="item.number"
            class="number-cell"
            :disabled="isLocked || isSubmitting"
            :class="{ 'number-cell--best': safestWingo.number === item.number }"
            @click="requestWinGo(item.number)"
          >
            <span class="number-cell__fill" :style="{ height: heatHeight(item.stake, maxWinGoNumberStake) }"></span>
            <span class="number-cell__n">{{ item.number }}</span>
            <span class="number-cell__stake">{{ formatViMoney(item.stake.toString()) }}</span>
          </button>
        </div>

        <button
          class="smart-btn"
          :disabled="isLocked || isSubmitting"
          @click="requestWinGo(safestWingo.number)"
        >
          ✨ One-Click Safest: Số {{ safestWingo.number }} ({{ formatViMoney(safestWingo.pl.toString()) }})
        </button>
      </section>

      <section v-else-if="room.game_type === 2" class="board board-k3">
        <div class="metric-grid metric-grid--four">
          <article v-for="card in k3PrimaryCards" :key="card.key" class="metric-card">
            <p class="metric-card__label">{{ card.label }}</p>
            <strong class="metric-card__value">{{ formatViMoney(card.stake.toString()) }}</strong>
            <span class="metric-card__meta">{{ card.players.toLocaleString('vi-VN') }} lượt</span>
          </article>
        </div>

        <div class="sum-grid">
          <button
            v-for="entry in k3Sums"
            :key="entry.sum"
            class="sum-cell"
            :disabled="isLocked || isSubmitting"
            @click="k3Dice = diceForSum(entry.sum)"
          >
            <span class="sum-cell__bg" :style="{ opacity: Math.max(0.12, Math.min(1, entry.stake / maxK3SumStake)) }"></span>
            <span class="sum-cell__num">{{ entry.sum }}</span>
            <span class="sum-cell__stake">{{ formatViMoney(entry.stake.toString()) }}</span>
          </button>
        </div>

        <div class="option-block">
          <div class="option-block__head">
            <h4>2 số trùng</h4>
          </div>
          <div class="mini-grid mini-grid--6">
            <article v-for="card in k3PairCards" :key="card.key" class="mini-card">
              <p class="mini-card__label">{{ card.label }}</p>
              <strong class="mini-card__stake">{{ formatViMoney(card.stake.toString()) }}</strong>
            </article>
          </div>
        </div>

        <div class="option-block">
          <div class="option-block__head">
            <h4>3 số trùng</h4>
          </div>
          <div class="mini-grid mini-grid--6">
            <article v-for="card in k3TripleCards" :key="card.key" class="mini-card mini-card--violet">
              <p class="mini-card__label">{{ card.label }}</p>
              <strong class="mini-card__stake">{{ formatViMoney(card.stake.toString()) }}</strong>
            </article>
          </div>
        </div>

        <div class="option-block">
          <div class="option-block__head">
            <h4>Khác số</h4>
          </div>
          <div class="mini-grid mini-grid--4">
            <article v-for="card in k3DiffCards" :key="card.key" class="mini-card mini-card--teal">
              <p class="mini-card__label">{{ card.label }}</p>
              <strong class="mini-card__stake">{{ formatViMoney(card.stake.toString()) }}</strong>
            </article>
          </div>
        </div>

        <div class="control-row">
          <label v-for="idx in 3" :key="idx" class="pick-col">
            <span>Xúc xắc {{ idx }}</span>
            <select v-model.number="k3Dice[idx - 1]" :disabled="isLocked || isSubmitting">
              <option v-for="v in 6" :key="v" :value="v">{{ v }}</option>
            </select>
          </label>
        </div>

        <div class="actions-row">
          <button class="smart-btn smart-btn--ghost" :disabled="isLocked || isSubmitting" @click="applySafestK3">
            ✨ Chọn tổng lợi nhuận cao nhất
          </button>
          <button class="smart-btn" :disabled="isLocked || isSubmitting" @click="requestK3">
            Submit: {{ k3Dice.join('-') }} (P/L {{ formatViMoney(calculateK3PL(sumDice(k3Dice)).toString()) }})
          </button>
        </div>
      </section>

      <section v-else class="board board-5d">
        <div class="matrix-wrap">
          <div v-for="(row, rowIdx) in lotteryMatrix" :key="`row-${rowIdx}`" class="matrix-row">
            <span class="matrix-label">{{ String.fromCharCode(65 + rowIdx) }}</span>
            <div class="matrix-cells">
              <button
                v-for="(cell, digit) in row"
                :key="cell.key"
                class="matrix-cell"
                :disabled="isLocked || isSubmitting"
                @click="lotteryDigits[rowIdx] = digit"
              >
                <span
                  class="matrix-cell__bg"
                  :style="{ opacity: Math.max(0.08, Math.min(1, cell.stake / (lotteryRowMax[rowIdx] || 1))) }"
                ></span>
                <span class="matrix-cell__digit">{{ digit }}</span>
                <span class="matrix-cell__stake">{{ formatViMoney(cell.stake.toString()) }}</span>
              </button>
            </div>
          </div>
        </div>

        <div class="option-block option-block--lottery-props">
          <div class="option-block__head">
            <h4>Cửa theo vị trí A-E</h4>
          </div>
          <div class="lottery-property-list">
            <div v-for="row in lotteryPositionPropertyRows" :key="row.position" class="lottery-property-row">
              <span class="lottery-property-row__label">{{ row.position }}</span>
              <div class="mini-grid mini-grid--4">
                <article v-for="card in row.items" :key="card.key" class="mini-card mini-card--slate">
                  <p class="mini-card__label">{{ compactOptionLabel(card.label) }}</p>
                  <strong class="mini-card__stake">{{ formatViMoney(card.stake.toString()) }}</strong>
                </article>
              </div>
            </div>
          </div>
        </div>

        <div class="option-block option-block--lottery-sum">
          <div class="option-block__head">
            <h4>SUM</h4>
          </div>
          <div class="mini-grid mini-grid--4">
            <article v-for="card in lotterySumPropertyCards" :key="card.key" class="mini-card mini-card--gold">
              <p class="mini-card__label">{{ compactOptionLabel(card.label) }}</p>
              <strong class="mini-card__stake">{{ formatViMoney(card.stake.toString()) }}</strong>
            </article>
          </div>
        </div>

        <div class="control-row control-row--5d">
          <label v-for="idx in 5" :key="idx" class="pick-col">
            <span>Vị trí {{ String.fromCharCode(64 + idx) }}</span>
            <select v-model.number="lotteryDigits[idx - 1]" :disabled="isLocked || isSubmitting">
              <option v-for="v in 10" :key="v - 1" :value="v - 1">{{ v - 1 }}</option>
            </select>
          </label>
        </div>

        <button class="smart-btn" :disabled="isLocked || isSubmitting" @click="requestLottery">
          Submit: {{ lotteryDigits.join('') }} (P/L {{ formatViMoney(calculateLotteryPL(lotteryDigits).toString()) }})
        </button>
      </section>

      <section v-if="optionStakeRows.length" class="option-summary">
        <div class="option-summary__head">
          <h4>Tổng cược theo cửa</h4>
          <strong>{{ formatViMoney(totalStake.toString()) }}</strong>
        </div>

        <div class="option-summary__grid">
          <article v-for="item in optionStakeRows" :key="item.key" class="option-summary__card">
            <p class="option-summary__label">{{ item.label }}</p>
            <strong class="option-summary__stake">{{ formatViMoney(item.stake.toString()) }}</strong>
            <span class="option-summary__meta">{{ item.players.toLocaleString('vi-VN') }} người cược</span>
          </article>
        </div>
      </section>
    </div>
  </article>
</template>

<style scoped>
.control-room-card {
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 22px;
  overflow: hidden;
  background: linear-gradient(180deg, rgba(9, 18, 38, 0.82), rgba(8, 13, 26, 0.92));
}

.room-head {
  padding: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.room-title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.room-title {
  font-size: 0.95rem;
  font-weight: 900;
  letter-spacing: 0.06em;
  color: #f8fafc;
}

.room-chip {
  padding: 3px 8px;
  border-radius: 999px;
  font-size: 0.63rem;
  font-weight: 800;
  letter-spacing: 0.06em;
}

.room-chip--open {
  background: rgba(16, 185, 129, 0.18);
  color: #a7f3d0;
}

.room-chip--lock {
  background: rgba(245, 158, 11, 0.16);
  color: #fde68a;
}

.room-period {
  margin-top: 3px;
  font-size: 0.75rem;
  color: #94a3b8;
}

.room-manual-state {
  margin-top: 6px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  border-radius: 999px;
  background: rgba(251, 113, 133, 0.16);
  color: #fecdd3;
  font-size: 1rem;
  font-weight: 800;
  letter-spacing: 0.03em;
}

.room-timer {
  text-align: right;
}

.room-timer__label {
  font-size: 0.62rem;
  letter-spacing: 0.08em;
  color: #94a3b8;
}

.room-timer__value {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 1.1rem;
  color: #f8fafc;
}

.room-body {
  padding: 14px;
}

.board {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.board-5d {
  display: grid;
  gap: 12px;
}

.metric-grid {
  display: grid;
  gap: 8px;
}

.metric-grid--four {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.board-k3 > .metric-grid--four,
.board-5d > .metric-grid--four {
  grid-template-columns: repeat(6, minmax(0, 1fr));
}

.board-k3 > .metric-grid--four .metric-card:first-child,
.board-5d > .metric-grid--four .metric-card:first-child {
  grid-column: span 2;
}

.metric-card {
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(15, 23, 42, 0.72);
  padding: 10px;
  display: grid;
  gap: 4px;
}

.metric-card__label {
  font-size: 0.66rem;
  font-weight: 800;
  color: #cbd5e1;
}

.metric-card__value {
  font-size: 0.8rem;
  font-weight: 900;
  color: #f8fafc;
}

.metric-card__meta {
  font-size: 0.58rem;
  font-weight: 700;
  color: #94a3b8;
}

.color-cards {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.color-card {
  border-radius: 12px;
  padding: 10px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  color: #f8fafc;
}

.color-card p {
  font-size: 0.68rem;
  font-weight: 800;
}

.color-card strong {
  display: block;
  margin-top: 2px;
  font-size: 0.76rem;
}

.color-card span {
  display: block;
  margin-top: 4px;
  font-size: 0.56rem;
  font-weight: 700;
  color: rgba(248, 250, 252, 0.82);
}

.color-card--green {
  background: linear-gradient(140deg, rgba(16, 185, 129, 0.35), rgba(6, 95, 70, 0.45));
}

.color-card--red {
  background: linear-gradient(140deg, rgba(251, 113, 133, 0.35), rgba(136, 19, 55, 0.45));
}

.color-card--violet {
  background: linear-gradient(140deg, rgba(168, 85, 247, 0.35), rgba(91, 33, 182, 0.45));
}

.numbers-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
}

.number-cell {
  position: relative;
  border-radius: 14px;
  min-height: 96px;
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid rgba(148, 163, 184, 0.2);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.number-cell__fill {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(180deg, rgba(20, 184, 166, 0.2), rgba(251, 113, 133, 0.7));
}

.number-cell__n,
.number-cell__stake {
  position: relative;
  z-index: 1;
}

.number-cell__n {
  font-size: 1rem;
  font-weight: 900;
  color: #f8fafc;
}

.number-cell__stake {
  font-size: 0.6rem;
  font-weight: 800;
  color: #fde68a;
}

.number-cell--best {
  outline: 2px solid rgba(16, 185, 129, 0.75);
}

.sum-grid {
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: 7px;
}

.sum-cell {
  position: relative;
  border-radius: 10px;
  min-height: 58px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(15, 23, 42, 0.85);
  overflow: hidden;
  display: grid;
  place-items: center;
}

.sum-cell__bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(140deg, #f59e0b, #fb7185);
}

.sum-cell__num,
.sum-cell__stake {
  position: relative;
  z-index: 1;
}

.sum-cell__num {
  font-size: 0.78rem;
  font-weight: 900;
  color: #f8fafc;
}

.sum-cell__stake {
  font-size: 0.6rem;
  color: #fde68a;
}

.option-block {
  display: grid;
  gap: 8px;
}

.option-block__head h4 {
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.04em;
  color: #cbd5e1;
}

.mini-grid {
  display: grid;
  gap: 6px;
}

.mini-grid--6 {
  grid-template-columns: repeat(6, minmax(0, 1fr));
}

.mini-grid--4 {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.mini-card {
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(15, 23, 42, 0.72);
  padding: 8px;
  display: grid;
  gap: 4px;
}

.mini-card--amber {
  background: linear-gradient(140deg, rgba(245, 158, 11, 0.22), rgba(15, 23, 42, 0.82));
}

.mini-card--violet {
  background: linear-gradient(140deg, rgba(168, 85, 247, 0.22), rgba(15, 23, 42, 0.82));
}

.mini-card--teal {
  background: linear-gradient(140deg, rgba(20, 184, 166, 0.2), rgba(15, 23, 42, 0.82));
}

.mini-card--gold {
  background: linear-gradient(140deg, rgba(250, 204, 21, 0.18), rgba(15, 23, 42, 0.82));
}

.mini-card--slate {
  background: rgba(30, 41, 59, 0.72);
}

.mini-card__label {
  font-size: 0.6rem;
  font-weight: 800;
  color: #cbd5e1;
}

.mini-card__stake {
  font-size: 0.7rem;
  font-weight: 900;
  color: #f8fafc;
}

.control-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.control-row--5d {
  grid-template-columns: repeat(5, minmax(0, 1fr));
}

.pick-col {
  display: grid;
  gap: 4px;
}

.pick-col span {
  font-size: 0.62rem;
  color: #94a3b8;
  font-weight: 700;
}

.pick-col select {
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(15, 23, 42, 0.8);
  color: #f8fafc;
  padding: 6px;
  font-size: 0.72rem;
}

.actions-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.smart-btn {
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: linear-gradient(120deg, #f97316, #ef4444);
  color: #fff;
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.02em;
  padding: 10px 12px;
}

.smart-btn--ghost {
  background: rgba(15, 23, 42, 0.7);
  color: #cbd5e1;
}

.smart-btn:disabled,
.number-cell:disabled,
.sum-cell:disabled,
.matrix-cell:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.matrix-wrap {
  display: grid;
  gap: 6px;
}

.lottery-property-list {
  display: grid;
  gap: 8px;
}

.lottery-property-row {
  display: grid;
  grid-template-columns: 24px 1fr;
  gap: 6px;
  align-items: start;
}

.lottery-property-row__label {
  padding-top: 8px;
  text-align: center;
  font-size: 0.68rem;
  font-weight: 900;
  color: #f8fafc;
}

.matrix-row {
  display: grid;
  grid-template-columns: 24px 1fr;
  gap: 6px;
  align-items: center;
}

.matrix-label {
  text-align: center;
  font-size: 0.66rem;
  font-weight: 900;
  color: #f8fafc;
}

.matrix-cells {
  display: grid;
  grid-template-columns: repeat(10, minmax(0, 1fr));
  gap: 4px;
}

.matrix-cell {
  position: relative;
  border-radius: 8px;
  min-height: 54px;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(15, 23, 42, 0.84);
  display: grid;
  place-items: center;
}

.matrix-cell__bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(140deg, #22c55e, #10b981);
}

.matrix-cell__digit,
.matrix-cell__stake {
  position: relative;
  z-index: 1;
}

.matrix-cell__digit {
  font-size: 0.72rem;
  color: #f8fafc;
  font-weight: 800;
}

.matrix-cell__stake {
  font-size: 0.56rem;
  color: #bbf7d0;
}

.option-summary {
  margin-top: 2px;
  display: grid;
  gap: 10px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding-top: 12px;
}

.option-summary__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.option-summary__head h4 {
  font-size: 0.76rem;
  font-weight: 900;
  color: #f8fafc;
}

.option-summary__head strong {
  font-size: 0.74rem;
  color: #67e8f9;
}

.option-summary__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.option-summary__card {
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(15, 23, 42, 0.72);
  padding: 10px;
  display: grid;
  gap: 3px;
}

.option-summary__label {
  font-size: 0.68rem;
  font-weight: 800;
  color: #e2e8f0;
  line-height: 1.35;
}

.option-summary__stake {
  font-size: 0.78rem;
  color: #fef08a;
}

.option-summary__meta {
  font-size: 0.62rem;
  color: #94a3b8;
}

.risk-alert {
  box-shadow: 0 0 0 1px rgba(251, 113, 133, 0.45), 0 0 22px rgba(244, 63, 94, 0.22);
}

.risk-safe {
  box-shadow: 0 0 0 1px rgba(16, 185, 129, 0.36), 0 0 20px rgba(16, 185, 129, 0.16);
}

.card-pulse {
  animation: roomPulse 1.2s ease-out;
}

@keyframes roomPulse {
  0% { transform: scale(1); filter: brightness(1); }
  25% { transform: scale(1.01); filter: brightness(1.1); }
  100% { transform: scale(1); filter: brightness(1); }
}

@media (max-width: 1280px) {
  .metric-grid--four {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .board-k3 > .metric-grid--four,
  .board-5d > .metric-grid--four {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
  .sum-grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
  .mini-grid--6 {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (min-width: 1380px) {
  .board-5d {
    grid-template-columns: minmax(0, 1.35fr) minmax(280px, 0.8fr);
    align-items: start;
  }

  .matrix-wrap {
    grid-column: 1;
  }

  .option-block--lottery-props {
    grid-column: 2;
    grid-row: 1 / span 2;
    align-self: stretch;
    align-content: start;
  }

  .option-block--lottery-sum {
    grid-column: 1;
  }

  .control-row--5d,
  .board-5d > .smart-btn {
    grid-column: 1 / -1;
  }
}

@media (max-width: 768px) {
  .metric-grid--four,
  .board-k3 > .metric-grid--four,
  .board-5d > .metric-grid--four {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    padding-bottom: 2px;
    scroll-snap-type: x proximity;
  }

  .metric-grid--four .metric-card,
  .board-k3 > .metric-grid--four .metric-card,
  .board-5d > .metric-grid--four .metric-card {
    min-width: 148px;
    flex: 0 0 148px;
    scroll-snap-align: start;
  }

  .board-k3 > .metric-grid--four .metric-card:first-child,
  .board-5d > .metric-grid--four .metric-card:first-child {
    grid-column: auto;
    min-width: 196px;
    flex-basis: 196px;
  }

  .numbers-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
  .actions-row {
    grid-template-columns: 1fr;
  }
  .control-row--5d {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
  .option-summary__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .mini-grid--4 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .lottery-property-row {
    grid-template-columns: 1fr;
  }
  .lottery-property-row__label {
    padding-top: 0;
    text-align: left;
  }
}
</style>
