<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed } from 'vue'
import { request, type ApiError } from '@/shared/api/http'
import { useAuthStore } from '@/stores/auth'
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

interface ManualResult {
  result: string
  big_small: string
  color: string
  payload: any
}

const auth = useAuthStore()
const rooms = ref<RoomStats[]>([])
const loading = ref(false) 
const error = ref('')
const activeTab = ref(1) // 1: WinGo, 2: K3, 3: Lottery
const autoRefresh = ref(true)
const serverTimeOffsetMs = ref(0)
const settingResult = ref<number | null>(null) // period id being edited

// K3/Lottery Selection State
const k3Selection = ref<Record<string, number[]>>({}) // roomCode -> [d1, d2, d3]
const lotterySelection = ref<Record<string, number[]>>({}) // roomCode -> [d1, d2, d3, d4, d5]

// Reactive clock tick to force countdown re-renders every second
const clockTick = ref(Date.now())

let timer: number | undefined
let clockTimer: number | undefined

async function fetchStats() {
  if (!auth.isAuthenticated || !autoRefresh.value) {
    loading.value = false
    return
  }
  
  if (rooms.value.length === 0) loading.value = true
  
  try {
    const response = await request<{server_time: string, rooms: RoomStats[]}>('GET', '/v1/admin/rooms/stats', { token: auth.accessToken })
    rooms.value = response.rooms
    
    // Calculate time offset
    const serverTime = new Date(response.server_time).getTime()
    const clientTime = Date.now()
    serverTimeOffsetMs.value = serverTime - clientTime
    
    console.log('[ControlPanel] Server Time Sync:', {
        server: response.server_time,
        offsetMs: serverTimeOffsetMs.value
    })
    
    error.value = ''
  } catch (err: any) {
    error.value = err.message || 'Lỗi tải dữ liệu admin'
    console.error('[ControlPanel] Fetch error:', err)
  } finally {
    loading.value = false
  }
}



function getCountdown(drawAt: string | undefined) {
  // Reference clockTick to make Vue track this function as a reactive dependency
  const _tick = clockTick.value
  
  if (!drawAt) return '00:00'
  
  const normalizedDate = drawAt.includes('T') ? drawAt : drawAt.replace(' ', 'T')
  const drawTime = new Date(normalizedDate).getTime()
  
  if (isNaN(drawTime)) return '--:--'
  
  // Use server-synchronized time
  const nowSynced = _tick + serverTimeOffsetMs.value
  const diff = drawTime - nowSynced
  
  if (diff <= 0) return '00:00'
  
  const secs = Math.floor(diff / 1000)
  const m = Math.floor(secs / 60)
  const s = secs % 60
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

onMounted(() => {
  void fetchStats()
  timer = window.setInterval(fetchStats, 2000)
  // Tick every second to re-render countdowns
  clockTimer = window.setInterval(() => { clockTick.value = Date.now() }, 1000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
  if (clockTimer) clearInterval(clockTimer)
})

function getOptionLabel(key: string) {
    const lower = key.toLowerCase()
    if (lower === 'big') return 'Lớn'
    if (lower === 'small') return 'Nhỏ'
    if (lower === 'odd') return 'Lẻ'
    if (lower === 'even') return 'Chẵn'
    if (lower.startsWith('number_')) return 'Số ' + lower.split('_')[1]
    return key
}

function getGroupedStats(stats: BetStat[]) {
    const safeStats = stats || []
    const bigSmall = safeStats.filter(s => s.option_type === 2)
    const numbers = safeStats.filter(s => s.option_type === 1)
    const colors = safeStats.filter(s => s.option_type === 4)
    const other = safeStats.filter(s => ![1,2,4].includes(s.option_type))
    return { bigSmall, numbers, colors, other }
}

function calculateWinGoPL(room: RoomStats, outcome: number) {
    if (!room.bet_stats || room.bet_stats.length === 0) return 0
    
    const bigSmall = outcome >= 5 ? 'big' : 'small'
    const oddEven = outcome % 2 === 0 ? 'even' : 'odd'
    const colorTags: string[] = []
    if (outcome === 0) colorTags.push('red', 'violet')
    else if (outcome === 5) colorTags.push('green', 'violet')
    else colorTags.push(outcome % 2 === 0 ? 'red' : 'green')

    const winTags = new Set([
        'number_' + outcome,
        bigSmall,
        oddEven,
        ...colorTags
    ])

    let totalStake = 0
    let totalPayout = 0

    for (const stat of room.bet_stats) {
        const stake = parseFloat(stat.total_stake)
        totalStake += stake
        if (winTags.has(stat.option_key.toLowerCase())) {
            totalPayout += stake * 1.98
        }
    }

    return totalStake - totalPayout
}

function calculateK3PL(room: RoomStats, sum: number) {
    if (!room.bet_stats || room.bet_stats.length === 0) return 0
    
    const bigSmall = sum >= 11 ? 'big' : 'small'
    const oddEven = sum % 2 === 0 ? 'even' : 'odd'
    const winTags = new Set([
        'sum_' + sum,
        bigSmall,
        oddEven
    ])

    let totalStake = 0
    let totalPayout = 0

    for (const stat of room.bet_stats) {
        const stake = parseFloat(stat.total_stake)
        totalStake += stake
        if (winTags.has(stat.option_key.toLowerCase())) {
            totalPayout += stake * 1.98
        }
    }

    return totalStake - totalPayout
}

function getK3Dice(roomCode: string) {
    if (!k3Selection.value[roomCode]) {
        k3Selection.value[roomCode] = [1, 1, 1]
    }
    return k3Selection.value[roomCode]
}

function setK3Die(roomCode: string, index: number, val: number) {
    const dice = getK3Dice(roomCode)
    dice[index] = val
}

function getLotteryDice(roomCode: string) {
    if (!lotterySelection.value[roomCode]) {
        lotterySelection.value[roomCode] = [0, 0, 0, 0, 0]
    }
    return lotterySelection.value[roomCode]
}

function setLotteryDigit(roomCode: string, index: number, val: number) {
    const dice = getLotteryDice(roomCode)
    dice[index] = val
}

function calculateLotteryPL(room: RoomStats, digits: number[]) {
    if (!room.bet_stats || room.bet_stats.length === 0) return 0
    
    const result = digits.join('')
    const sum = digits.reduce((a, b) => a + b, 0)
    const last = digits[4]
    const bigSmall = last >= 5 ? 'big' : 'small'
    const oddEven = last % 2 === 0 ? 'even' : 'odd'

    const winTags = new Set([
        'pick5_' + result,
        'sum_' + sum,
        'last_' + last,
        bigSmall,
        oddEven
    ])

    let totalStake = 0
    let totalPayout = 0

    for (const stat of room.bet_stats) {
        const stake = parseFloat(stat.total_stake)
        totalStake += stake
        if (winTags.has(stat.option_key.toLowerCase())) {
            totalPayout += stake * 1.98
        }
    }

    return totalStake - totalPayout
}

async function setLotteryResult(room: RoomStats) {
    const digits = getLotteryDice(room.code)
    const result = digits.join('')
    const sum = digits.reduce((a, b) => a + b, 0)
    const last = digits[4]
    const bigSmall = last >= 5 ? 'big' : 'small'
    const oddEven = last % 2 === 0 ? 'even' : 'odd'

    const payload = {
        game_type: 'lottery',
        digits,
        sum,
        last_digit: last,
        result,
        big_small: bigSmall,
        odd_even: oddEven,
        tags: [
            'pick5_' + result,
            'sum_' + sum,
            'last_' + last,
            bigSmall,
            oddEven
        ],
        generated_at: new Date().toISOString()
    }

    await setManualResult(room.period!.id, room.code, room.game_type, result, bigSmall, '-', payload)
}

async function setManualResult(periodId: number, roomCode: string, gameType: number, result: string, bigSmall: string = '', color: string = '', payloadOverride: any = null) {
  if (!confirm(`Xác nhận đặt kết quả "${result}" cho phiên này?`)) return
  
  let payload: any = payloadOverride
  
  if (!payload) {
      payload = { game_type: gameType === 1 ? 'wingo' : gameType === 2 ? 'k3' : 'lottery' }
      if (gameType === 1) {
        const num = parseInt(result)
        bigSmall = num >= 5 ? 'big' : 'small'
        if (num === 0) color = 'red_violet'
        else if (num === 5) color = 'green_violet'
        else color = num % 2 === 0 ? 'red' : 'green'
        
        payload = {
            ...payload,
            number: num,
            result: result,
            big_small: bigSmall,
            color: color,
            tags: ["number_" + result, bigSmall, num % 2 === 0 ? 'even' : 'odd'], 
        }
        if (num === 0) payload.tags.push('red', 'violet')
        else if (num === 5) payload.tags.push('green', 'violet')
        else payload.tags.push(color)
      }
  }

  try {
    await request('POST', `/v1/admin/periods/${periodId}/result`, {
      token: auth.accessToken,
      body: {
        result,
        big_small: bigSmall,
        color,
        payload
      }
    })
    alert('Đã đặt kết quả thành công')
    void fetchStats()
  } catch (err: any) {
    alert('Lỗi: ' + err.message)
  }
}

function getParsedManualResult(room: RoomStats): ManualResult | null {
    if (!room.period?.manual_result) return null
    try {
        return JSON.parse(room.period.manual_result)
    } catch {
        return null
    }
}

const filteredRooms = computed(() => {
    console.log('[ControlPanel] Raw Rooms:', rooms.value)
    console.log('[ControlPanel] Active Tab:', activeTab.value)
    const filtered = rooms.value.filter(r => {
        return Number(r.game_type) === Number(activeTab.value)
    })
    console.log('[ControlPanel] Filtered Result:', filtered.length, 'rooms')
    return filtered
})


</script>

<template>
  <div class="admin-control min-h-screen bg-[#0f172a] text-slate-200 p-4 pb-20">
    <header class="mb-8 p-4 bg-slate-800/30 rounded-2xl border border-slate-700/50">
      <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h1 class="text-2xl font-black bg-gradient-to-r from-amber-400 via-orange-500 to-rose-500 bg-clip-text text-transparent uppercase tracking-tighter">
            Control Panel Engine
          </h1>
          <p class="text-slate-400 text-xs font-medium tracking-widest uppercase mt-1 opacity-70">Sổ lệnh cược & Can thiệp kết quả</p>
        </div>
        <div class="flex bg-slate-900/80 p-1 rounded-xl border border-slate-700">
           <button 
             v-for="(name, id) in {1: 'WinGo', 2: 'K3 Xúc Xắc', 3: '5D Lottery'}" :key="id"
             @click="activeTab = Number(id)"
             class="px-4 py-2 rounded-lg text-sm font-bold transition-all"
             :class="activeTab === Number(id) ? 'bg-amber-500 text-white shadow-lg' : 'text-slate-500 hover:text-slate-300'"
           >
             {{ name }}
           </button>
        </div>
        <div class="flex items-center gap-3 bg-slate-900/50 px-4 py-2 rounded-xl border border-slate-700/50">
           <span class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Tự động cập nhật</span>
           <button 
             @click="autoRefresh = !autoRefresh"
             class="w-10 h-5 rounded-full relative transition-colors"
             :class="autoRefresh ? 'bg-green-500/50' : 'bg-slate-700'"
           >
             <div class="absolute top-1 w-3 h-3 rounded-full bg-white transition-all shadow-sm"
                  :class="autoRefresh ? 'left-6 bg-green-200' : 'left-1'"></div>
           </button>
        </div>
      </div>
      <div v-if="error" class="mt-4 bg-red-500/20 text-red-400 px-4 py-2 rounded-lg border border-red-500/30 text-sm">
        {{ error }}
      </div>
    </header>

    <div v-if="loading && rooms.length === 0" class="flex flex-col items-center justify-center py-32 grayscale opacity-50">
        <div class="w-12 h-12 border-4 border-amber-500/20 border-t-amber-500 rounded-full animate-spin mb-4"></div>
        <p class="text-slate-400 font-medium animate-pulse uppercase tracking-widest text-xs">Đang đồng bộ dữ liệu...</p>
    </div>

    <div v-else-if="filteredRooms.length === 0" class="flex flex-col items-center justify-center py-32 bg-slate-800/20 rounded-3xl border border-dashed border-slate-700/50">
        <div class="text-slate-700 mb-6">
            <svg class="w-20 h-20 mx-auto opacity-20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
            </svg>
        </div>
        <h3 class="text-slate-400 font-bold uppercase tracking-[0.3em] text-sm mb-2">Không tìm thấy phòng chơi</h3>
        <p class="text-slate-600 text-xs italic">Vui lòng kiểm tra trạng thái Engine hoặc cấu hình phòng trong Database</p>
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div v-for="room in filteredRooms" :key="room.code" 
           class="bg-[#1e293b] border border-slate-700/50 rounded-2xl overflow-hidden shadow-2xl transition-all hover:border-slate-600">
        
        <!-- Room Header -->
        <div class="p-4 bg-slate-800/80 flex justify-between items-center border-b border-slate-700">
          <div>
            <div class="flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                <span class="text-lg font-black text-white uppercase tracking-tight">{{ room.code.replace('_', ' ') }}</span>
                <span v-if="getParsedManualResult(room)" 
                      class="text-[9px] bg-rose-500/20 text-rose-400 border border-rose-500/30 px-2 py-0.5 rounded-full font-bold uppercase animate-pulse">
                    Đã cài kết quả: {{ getParsedManualResult(room)?.result }}
                </span>
                <span v-else-if="!room.period || !room.period.id" 
                      class="text-[9px] bg-slate-700 text-slate-500 px-2 py-0.5 rounded-full font-bold uppercase">
                    Chờ kỳ mới...
                </span>
            </div>
            <div class="text-xs text-slate-400 mt-1">
              Kỳ hiện tại: <span class="text-amber-400 font-mono font-bold">{{ room.period && room.period.id ? room.period.period_no : '---' }}</span>
            </div>
          </div>
          <div class="text-right">
            <div class="text-2xl font-mono font-black text-orange-500 tabular-nums">
              {{ room.period && room.period.id ? getCountdown(room.period.draw_at) : '00:00' }}
            </div>
            <div class="text-[10px] uppercase tracking-widest font-bold text-slate-500 mt-0.5">Thời gian còn lại</div>
          </div>
        </div>

        <div class="p-4 space-y-6">
          <!-- Stats Summary -->
          <div v-if="room.bet_stats && room.bet_stats.length > 0" class="space-y-4">
             <div v-for="group in [getGroupedStats(room.bet_stats)]" :key="room.code + 'stats'" class="space-y-3">
                <!-- Big/Small Stats -->
                <div v-if="group.bigSmall.length" class="bg-slate-900/30 p-3 rounded-xl border border-slate-800/50">
                    <div class="flex justify-between items-center mb-3">
                        <h3 class="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em]">Tài / Xỉu</h3>
                        <span class="text-[10px] bg-slate-800 px-2 py-0.5 rounded text-slate-400">Real-time</span>
                    </div>
                    <div class="grid grid-cols-2 gap-3">
                        <div v-for="s in group.bigSmall" :key="s.option_key" 
                             class="flex justify-between items-center p-2 bg-slate-800/40 rounded-lg border border-slate-700/30">
                            <div>
                                <span class="text-xs font-bold block text-slate-300">{{ getOptionLabel(s.option_key) }}</span>
                                <span class="text-[10px] text-slate-500">{{ s.player_count }} lệnh</span>
                            </div>
                            <span class="text-sm font-black text-amber-500">{{ formatViMoney(s.total_stake) }}</span>
                        </div>
                    </div>
                </div>

                <!-- Number Stats -->
                <div v-if="group.numbers.length" class="bg-slate-900/30 p-3 rounded-xl border border-slate-800/50">
                    <h3 class="text-[10px] font-bold text-slate-500 mb-3 uppercase tracking-[0.2em]">Các con số</h3>
                    <div class="flex flex-wrap gap-2">
                        <div v-for="s in group.numbers" :key="s.option_key" 
                             class="flex items-center gap-2 bg-slate-800/60 px-2.5 py-1.5 rounded-lg border border-slate-700/50">
                            <span class="text-xs font-black text-slate-400">{{ s.option_key.split('_')[1] }}:</span>
                            <span class="text-xs font-bold text-amber-500">{{ formatViMoney(s.total_stake) }}</span>
                        </div>
                    </div>
                </div>
             </div>
          </div>
          <div v-else class="py-12 text-center">
            <div class="text-slate-600 mb-2">
                <svg class="w-12 h-12 mx-auto opacity-20" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-9 0h6"></path></svg>
            </div>
            <p class="text-slate-500 italic text-sm">Chưa có dữ liệu đặt cược</p>
          </div>

          <!-- Intervention Controls -->
          <div v-if="room.period && room.game_type === 1" class="border-t border-slate-700/50 pt-5">
            <div class="flex justify-between items-end mb-4">
                <h3 class="text-[10px] font-black text-orange-400 uppercase tracking-[0.2em]">Dự báo lợi nhuận WinGo</h3>
                <span class="text-[9px] text-slate-500 italic">* Lợi nhuận = Tổng cược - Tổng trả thưởng</span>
            </div>
            <div class="grid grid-cols-5 gap-2">
              <div v-for="n in 10" :key="n-1" class="flex flex-col gap-1 items-center">
                <button 
                  @click="setManualResult(room.period!.id, room.code, room.game_type, (n-1).toString())"
                  class="w-full aspect-square rounded-xl flex items-center justify-center font-black text-lg text-white transition-all hover:scale-105 active:scale-95 shadow-xl relative group"
                  :class="[
                    (n-1) === 0 ? 'bg-gradient-to-br from-red-500 to-purple-600' :
                    (n-1) === 5 ? 'bg-gradient-to-br from-green-500 to-purple-600' :
                    (n-1) % 2 === 0 ? 'bg-rose-500' : 'bg-emerald-500'
                  ]"
                >
                  {{ n-1 }}
                  <div class="absolute inset-0 bg-white/10 opacity-0 group-hover:opacity-100 rounded-xl transition-opacity"></div>
                </button>
                <div class="text-[9px] font-bold text-center w-full truncate"
                     :class="calculateWinGoPL(room, n-1) >= 0 ? 'text-green-400' : 'text-rose-400'">
                    {{ calculateWinGoPL(room, n-1) > 0 ? '+' : '' }}{{ formatViMoney(calculateWinGoPL(room, n-1).toString()) }}
                </div>
              </div>
            </div>
          </div>

          <!-- K3 Controls -->
          <div v-else-if="room.period && room.game_type === 2" class="border-t border-slate-700/50 pt-5">
             <div class="flex justify-between items-end mb-4">
                <h3 class="text-[10px] font-black text-amber-400 uppercase tracking-[0.2em]">Can thiệp K3 Xúc Xắc</h3>
                <div v-for="dice in [getK3Dice(room.code)]" :key="room.code" 
                     class="text-[10px] font-bold"
                     :class="calculateK3PL(room, dice[0]+dice[1]+dice[2]) >= 0 ? 'text-green-400' : 'text-rose-400'">
                    P/L: {{ formatViMoney(calculateK3PL(room, dice[0]+dice[1]+dice[2]).toString()) }}
                </div>
             </div>
             
             <div class="flex flex-col gap-4">
                <div class="flex justify-center gap-6">
                    <div v-for="(die, idx) in [0, 1, 2]" :key="idx" class="flex flex-col items-center gap-2">
                        <div class="text-[9px] uppercase text-slate-500 font-bold">Xí ngầu {{ idx+1 }}</div>
                        <select 
                            :value="getK3Dice(room.code)[idx]"
                            @change="(e: any) => setK3Die(room.code, idx, parseInt(e.target.value))"
                            class="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1 text-white font-bold text-sm focus:ring-1 ring-amber-500 outline-none"
                        >
                            <option v-for="v in 6" :key="v" :value="v">{{ v }}</option>
                        </select>
                    </div>
                </div>

                <div class="flex justify-center">
                    <button 
                        @click="setManualResult(room.period!.id, room.code, room.game_type, getK3Dice(room.code).join('-'))"
                        class="w-full py-3 bg-gradient-to-r from-amber-500 to-orange-600 rounded-xl font-black text-sm text-white shadow-lg hover:shadow-amber-500/20 active:scale-[0.98] transition-all uppercase tracking-widest"
                    >
                        Xác nhận kết quả: {{ getK3Dice(room.code).join(' - ') }}
                    </button>
                </div>
             </div>
          </div>

          <div v-else-if="room.period && room.game_type === 3" class="border-t border-slate-700/50 pt-5">
             <div class="flex justify-between items-end mb-4">
                <h3 class="text-[10px] font-black text-rose-400 uppercase tracking-[0.2em]">Can thiệp 5D Lottery</h3>
                <div v-for="dice in [getLotteryDice(room.code)]" :key="room.code" 
                     class="text-[10px] font-bold"
                     :class="calculateLotteryPL(room, dice) >= 0 ? 'text-green-400' : 'text-rose-400'">
                    Dự báo P/L: {{ formatViMoney(calculateLotteryPL(room, dice).toString()) }}
                </div>
             </div>
             
             <div class="flex flex-col gap-4">
                <div class="grid grid-cols-5 gap-4">
                    <div v-for="(die, idx) in [0, 1, 2, 3, 4]" :key="idx" class="flex flex-col items-center gap-2">
                        <div class="text-[9px] uppercase text-slate-500 font-bold">Vị trí {{ String.fromCharCode(65 + idx) }}</div>
                        <select 
                            :value="getLotteryDice(room.code)[idx]"
                            @change="(e: any) => setLotteryDigit(room.code, idx, parseInt(e.target.value))"
                            class="w-full bg-slate-800 border border-slate-700 rounded-lg px-1 py-1 text-white font-bold text-sm focus:ring-1 ring-rose-500 outline-none"
                        >
                            <option v-for="v in 10" :key="v-1" :value="v-1">{{ v-1 }}</option>
                        </select>
                    </div>
                </div>

                <div class="flex justify-center">
                    <button 
                        @click="setLotteryResult(room)"
                        class="w-full py-3 bg-gradient-to-r from-rose-500 to-pink-600 rounded-xl font-black text-sm text-white shadow-lg hover:shadow-rose-500/20 active:scale-[0.98] transition-all uppercase tracking-widest"
                    >
                        Xác nhận kết quả 5D: {{ getLotteryDice(room.code).join('') }}
                    </button>
                </div>
             </div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<style scoped>
.admin-control {
  font-family: 'Inter', sans-serif;
}
</style>
