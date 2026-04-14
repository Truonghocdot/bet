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
}>()

const emit = defineEmits<{
  (e: 'request-submit', payload: SubmitPayload): void
}>()

const k3Dice = ref<[number, number, number]>([1, 1, 1])
const lotteryDigits = ref<[number, number, number, number, number]>([0, 0, 0, 0, 0])

const stats = computed(() => props.room.bet_stats ?? [])

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

function stakeByPrefix(prefix: string): number {
  const lowerPrefix = prefix.toLowerCase()
  return stats.value
    .filter((item) => item.option_key.toLowerCase().startsWith(lowerPrefix))
    .reduce((sum, item) => sum + parseStake(item.total_stake), 0)
}

const totalStake = computed(() =>
  stats.value.reduce((sum, item) => sum + parseStake(item.total_stake), 0),
)

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
  if (period.status >= 3) return true
  return leftSeconds.value <= 5
})

const lockMessage = computed(() => {
  if (!props.room.period) return 'Chờ kỳ mới'
  if (props.room.period.status >= 3) return 'Đã khóa lệnh'
  if (leftSeconds.value <= 5) return `Khóa sau ${leftSeconds.value}s`
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

const bigStake = computed(() => stakeByKey('big'))
const smallStake = computed(() => stakeByKey('small'))
const oddStake = computed(() => stakeByKey('odd'))
const evenStake = computed(() => stakeByKey('even'))

function ratio(a: number, b: number): number {
  const total = a + b
  if (total <= 0) return 0.5
  return a / total
}

const bigSmallRatio = computed(() => ratio(bigStake.value, smallStake.value))
const oddEvenRatio = computed(() => ratio(oddStake.value, evenStake.value))

const colorCards = computed(() => [
  { key: 'green', label: 'Xanh', stake: stakeByKey('green'), players: stakeByPrefix('green_') },
  { key: 'red', label: 'Đỏ', stake: stakeByKey('red'), players: stakeByPrefix('red_') },
  { key: 'violet', label: 'Tím', stake: stakeByKey('violet'), players: stakeByPrefix('violet_') },
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

  for (const stat of stats.value) {
    if (winTags.has(stat.option_key.toLowerCase())) {
      payout += parseStake(stat.total_stake) * 1.98
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

function calculateK3PL(sum: number): number {
  const bigSmall = sum >= 11 ? 'big' : 'small'
  const oddEven = sum % 2 === 0 ? 'even' : 'odd'
  const winTags = new Set<string>([`sum_${sum}`, bigSmall, oddEven])
  let payout = 0
  for (const stat of stats.value) {
    if (winTags.has(stat.option_key.toLowerCase())) {
      payout += parseStake(stat.total_stake) * 1.98
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
  const dice = k3Dice.value
  const sum = sumDice(dice)
  const bigSmall = sum >= 11 ? 'big' : 'small'
  const oddEven = sum % 2 === 0 ? 'even' : 'odd'
  const result = `${dice[0]}-${dice[1]}-${dice[2]}`
  emit('request-submit', {
    roomCode: props.room.code,
    periodId,
    result,
    bigSmall,
    color: '-',
    payload: {
      game_type: 'k3',
      dice,
      sum,
      big_small: bigSmall,
      odd_even: oddEven,
      tags: [`sum_${sum}`, bigSmall, oddEven],
      generated_at: new Date().toISOString(),
    },
    title: `K3 ${result}`,
  })
}

const otherK3Stats = computed(() =>
  stats.value.filter((item) => !item.option_key.toLowerCase().startsWith('sum_')).slice(0, 12),
)

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

function calculateLotteryPL(digits: number[]): number {
  const result = digits.join('')
  const sum = digits.reduce((a, b) => a + b, 0)
  const last = digits[4] ?? 0
  const bigSmall = last >= 5 ? 'big' : 'small'
  const oddEven = last % 2 === 0 ? 'even' : 'odd'
  const winTags = new Set<string>([`pick5_${result}`, `sum_${sum}`, `last_${last}`, bigSmall, oddEven])

  let payout = 0
  for (const stat of stats.value) {
    if (winTags.has(stat.option_key.toLowerCase())) {
      payout += parseStake(stat.total_stake) * 1.98
    }
  }
  return totalStake.value - payout
}

function requestLottery(): void {
  const periodId = props.room.period?.id
  if (!periodId || isLocked.value || props.isSubmitting) return
  const digits = lotteryDigits.value
  const result = digits.join('')
  const sum = digits.reduce((a, b) => a + b, 0)
  const last = digits[4] ?? 0
  const bigSmall = last >= 5 ? 'big' : 'small'
  const oddEven = last % 2 === 0 ? 'even' : 'odd'
  emit('request-submit', {
    roomCode: props.room.code,
    periodId,
    result,
    bigSmall,
    color: '-',
    payload: {
      game_type: 'lottery',
      digits,
      sum,
      last_digit: last,
      result,
      big_small: bigSmall,
      odd_even: oddEven,
      tags: [`pick5_${result}`, `sum_${sum}`, `last_${last}`, bigSmall, oddEven],
      generated_at: new Date().toISOString(),
    },
    title: `5D ${result}`,
  })
}

const previewPL = computed(() => {
  if (props.room.game_type === 1) return calculateWinGoPL(safestWingo.value.number)
  if (props.room.game_type === 2) return calculateK3PL(sumDice(k3Dice.value))
  return calculateLotteryPL(lotteryDigits.value)
})

const riskClass = computed(() => (previewPL.value < 0 ? 'risk-alert' : 'risk-safe'))

function plClass(value: number): string {
  return value >= 0 ? 'text-emerald-300' : 'text-rose-300'
}

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
          <span v-if="manualResult?.result" class="room-chip room-chip--manual">Đã cài: {{ manualResult.result }}</span>
        </div>
        <p class="room-period">Kỳ: {{ room.period?.period_no ?? '---' }}</p>
      </div>
      <div class="room-timer">
        <p class="room-timer__label">Time Left</p>
        <strong class="room-timer__value">{{ countdownLabel }}</strong>
      </div>
    </header>

    <div class="room-body">
      <section v-if="room.game_type === 1" class="board board-wingo">
        <div class="duel-grid">
          <div class="duel-row">
            <div class="duel-label"><span>Lớn</span><span>{{ formatViMoney(bigStake.toString()) }}</span></div>
            <div class="duel-track">
              <div class="duel-fill duel-fill--emerald" :style="{ width: `${Math.round(bigSmallRatio * 100)}%` }"></div>
            </div>
            <div class="duel-label"><span>Nhỏ</span><span>{{ formatViMoney(smallStake.toString()) }}</span></div>
          </div>
          <div class="duel-row">
            <div class="duel-label"><span>Lẻ</span><span>{{ formatViMoney(oddStake.toString()) }}</span></div>
            <div class="duel-track">
              <div class="duel-fill duel-fill--amber" :style="{ width: `${Math.round(oddEvenRatio * 100)}%` }"></div>
            </div>
            <div class="duel-label"><span>Chẵn</span><span>{{ formatViMoney(evenStake.toString()) }}</span></div>
          </div>
        </div>

        <div class="color-cards">
          <div v-for="card in colorCards" :key="card.key" class="color-card" :class="`color-card--${card.key}`">
            <p>{{ card.label }}</p>
            <strong>{{ formatViMoney(card.stake.toString()) }}</strong>
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
            <span class="number-cell__pl" :class="plClass(calculateWinGoPL(item.number))">
              {{ formatViMoney(calculateWinGoPL(item.number).toString()) }}
            </span>
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

        <div class="dice-badges">
          <span v-for="item in otherK3Stats" :key="item.option_key" class="dice-badge">
            {{ item.option_key }} · {{ formatViMoney(item.total_stake) }}
          </span>
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

.room-chip--manual {
  background: rgba(251, 113, 133, 0.16);
  color: #fecdd3;
}

.room-period {
  margin-top: 3px;
  font-size: 0.75rem;
  color: #94a3b8;
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

.duel-grid {
  display: grid;
  gap: 8px;
}

.duel-row {
  display: grid;
  grid-template-columns: minmax(52px, auto) 1fr minmax(52px, auto);
  gap: 8px;
  align-items: center;
}

.duel-label {
  display: grid;
  gap: 2px;
  font-size: 0.66rem;
  font-weight: 700;
  color: #cbd5e1;
}

.duel-track {
  height: 10px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.2);
  overflow: hidden;
}

.duel-fill {
  height: 100%;
}

.duel-fill--emerald {
  background: linear-gradient(90deg, #10b981, #22d3ee);
}

.duel-fill--amber {
  background: linear-gradient(90deg, #f59e0b, #fb7185);
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
.number-cell__pl {
  position: relative;
  z-index: 1;
}

.number-cell__n {
  font-size: 1rem;
  font-weight: 900;
  color: #f8fafc;
}

.number-cell__pl {
  font-size: 0.63rem;
  font-weight: 700;
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

.dice-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.dice-badge {
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  padding: 4px 8px;
  font-size: 0.62rem;
  color: #cbd5e1;
  background: rgba(15, 23, 42, 0.7);
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
  .sum-grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .numbers-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
  .actions-row {
    grid-template-columns: 1fr;
  }
  .control-row--5d {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
