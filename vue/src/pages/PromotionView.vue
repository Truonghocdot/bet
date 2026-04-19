<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { request, type ApiError } from '@/shared/api/http'
import type { ContentListResponse, ContentNewsItem, ManagedAffiliateUser } from '@/shared/api/types'
import { useAuthStore } from '@/stores/auth'

const activeTab = ref<'affiliate' | 'promotion' | 'news'>('affiliate')
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')
const promotionItems = ref<ContentNewsItem[]>([])
const newsItems = ref<ContentNewsItem[]>([])
const invitedUsersCount = ref(0)
const managedUsers = ref<ManagedAffiliateUser[]>([])
const copied = ref(false)

const isBecomeAgencyOpen = ref(false)
const staffInviteCode = ref('')
const becomeAgencyLoading = ref(false)
const becomeAgencyError = ref('')

const isClient = computed(() => auth.user?.role === 2)
const isAgency = computed(() => auth.user?.role === 4)

const affiliateStatusLabel = computed(() => {
  const status = auth.affiliateProfile?.status
  if (status === 2) return 'Đang hoạt động'
  if (status === 1) return 'Đang chờ duyệt'
  if (status === 3) return 'Tạm khóa'
  return 'Chưa đăng ký'
})

const affiliateCards = computed(() => [
  { label: 'Mã giới thiệu', value: auth.affiliateProfile?.ref_code || '—' },
  { label: 'Người đã mời', value: invitedUsersCount.value.toString() },
  { label: 'Link giới thiệu', value: auth.affiliateProfile?.ref_link || '—' },
  { label: 'Trạng thái', value: affiliateStatusLabel.value },
])

const tabTitle = computed(() => {
  if (activeTab.value === 'affiliate') return 'Affiliate'
  if (activeTab.value === 'promotion') return 'Khuyến mãi'
  return 'Tin tức'
})

async function loadContent() {
  loading.value = true
  error.value = ''
  try {
    const [promotions, news] = await Promise.all([
      request<ContentListResponse>('GET', '/v1/content/promotions?page=1&page_size=50'),
      request<ContentListResponse>('GET', '/v1/content/news?page=1&page_size=50'),
    ])
    promotionItems.value = promotions.items || []
    newsItems.value = news.items || []
  } catch (e: any) {
    const err = e as ApiError
    error.value = err?.message ?? 'Không thể tải dữ liệu nội dung'
    promotionItems.value = []
    newsItems.value = []
  } finally {
    loading.value = false
  }
}

async function loadAffiliateStats() {
  if (!auth.isAuthenticated) return
  try {
    const res = await request<{ invited_users_count: number }>('GET', '/v1/affiliate/summary', {
      token: auth.accessToken,
    })
    invitedUsersCount.value = Number(res.invited_users_count ?? 0)
  } catch {
    invitedUsersCount.value = 0
  }
}

async function loadManagedUsers() {
  if (!auth.isAuthenticated || !isAgency.value) {
    managedUsers.value = []
    return
  }

  try {
    const res = await request<{ items: ManagedAffiliateUser[] }>('GET', '/v1/affiliate/managed-users', {
      token: auth.accessToken,
    })
    managedUsers.value = res.items || []
  } catch {
    managedUsers.value = []
  }
}

function affiliateReferralStatusLabel(status: number): string {
  if (status === 2) return 'Đã nạp đầu'
  if (status === 1) return 'Chờ nạp đầu'
  if (status === 3) return 'Không hợp lệ'
  return '—'
}

async function copyReferralLink() {
  const link = auth.affiliateProfile?.ref_link
  if (!link) return
  try {
    await navigator.clipboard.writeText(link)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch {
    copied.value = false
  }
}

function openBecomeAgency() {
  becomeAgencyError.value = ''
  staffInviteCode.value = ''
  isBecomeAgencyOpen.value = true
}

async function submitBecomeAgency() {
  if (!auth.accessToken) return
  becomeAgencyError.value = ''
  becomeAgencyLoading.value = true
  try {
    const res = await request<any>('POST', '/v1/affiliate/become-agency', {
      token: auth.accessToken,
      body: { staff_ref_code: staffInviteCode.value.trim() },
    })
    auth.applyAuthResponse(res)
    await loadAffiliateStats()
    isBecomeAgencyOpen.value = false
  } catch (e: any) {
    const err = e as ApiError
    becomeAgencyError.value = err?.message ?? 'Không thể nâng cấp tài khoản'
  } finally {
    becomeAgencyLoading.value = false
  }
}

onMounted(async () => {
  if (auth.isAuthenticated && !auth.affiliateProfile) {
    try {
      await auth.fetchMe()
    } catch {
      // handled by auth store
    }
  }
  await loadAffiliateStats()
  await loadManagedUsers()
  await loadContent()
})
</script>

<template>
  <div class="space-y-3 pb-8">
    <div class="flex items-center gap-2 border-b border-slate-100 bg-white px-4 py-4">
      <span class="material-symbols-outlined text-[1.3rem] text-primary">group</span>
      <h1 class="text-[1rem] font-black text-on-surface">{{ tabTitle }}</h1>
    </div>

    <section class="px-3">
      <div class="grid grid-cols-3 gap-2 rounded-[18px] bg-white p-1.5 shadow-sm border border-slate-100">
        <button
          class="min-h-10 rounded-[12px] text-[0.78rem] font-black transition-all"
          :class="activeTab === 'affiliate' ? 'bg-primary text-white' : 'text-slate-500'"
          @click="activeTab = 'affiliate'"
        >
          Affiliate
        </button>
        <button
          class="min-h-10 rounded-[12px] text-[0.78rem] font-black transition-all"
          :class="activeTab === 'promotion' ? 'bg-primary text-white' : 'text-slate-500'"
          @click="activeTab = 'promotion'"
        >
          Khuyến mãi
        </button>
        <button
          class="min-h-10 rounded-[12px] text-[0.78rem] font-black transition-all"
          :class="activeTab === 'news' ? 'bg-primary text-white' : 'text-slate-500'"
          @click="activeTab = 'news'"
        >
          Tin tức
        </button>
      </div>
    </section>

    <section v-if="activeTab === 'affiliate'" class="flex flex-col gap-3 px-3">
      <div class="rounded-[18px] bg-gradient-to-br from-[#ff6d66] to-[#ff9f98] p-4 text-white">
        <p class="m-0 text-[0.7rem] uppercase tracking-[0.08em] text-white/80">
          {{ isAgency ? 'Khu vực Đại lý' : 'Chương trình Affiliate' }}
        </p>
        <h2 class="mt-2 text-[1.15rem] font-black">
          {{ isAgency ? 'Thống kê đại lý' : 'Mời bạn bè, nhận hoa hồng theo doanh thu' }}
        </h2>
        <p class="mt-2 text-[0.78rem] leading-5 text-white/90">
          {{ isAgency ? 'Theo dõi người đã mời và hiệu quả giới thiệu của bạn.' : 'Chia sẻ mã giới thiệu, theo dõi người đăng ký và nhận thưởng theo chính sách đại lý.' }}
        </p>
      </div>

      <div class="grid gap-2 md:grid-cols-3">
        <article v-for="item in affiliateCards" :key="item.label" class="rounded-[16px] border border-slate-100 bg-white p-4 shadow-sm">
          <p class="m-0 text-[0.7rem] font-bold uppercase tracking-[0.05em] text-slate-500">{{ item.label }}</p>
          <p class="mt-2 break-all text-[0.9rem] font-black text-on-surface">{{ item.value }}</p>
        </article>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <button
          type="button"
          class="inline-flex w-fit items-center gap-2 rounded-full bg-primary px-4 py-2 text-[0.75rem] font-black text-white"
          @click="copyReferralLink"
        >
          <span class="material-symbols-outlined text-[1rem]">content_copy</span>
          {{ copied ? 'Đã copy link' : 'Copy link giới thiệu' }}
        </button>

        <button
          v-if="isClient"
          type="button"
          class="inline-flex w-fit items-center gap-2 rounded-full bg-white px-4 py-2 text-[0.75rem] font-black text-primary border border-rose-200"
          @click="openBecomeAgency"
        >
          <span class="material-symbols-outlined text-[1rem]">workspace_premium</span>
          Trở thành đại lý
        </button>
      </div>

      <div v-if="isAgency" class="overflow-hidden rounded-[18px] border border-slate-100 bg-white shadow-sm">
        <div class="flex items-center justify-between border-b border-slate-100 px-4 py-3">
          <div>
            <p class="m-0 text-[0.72rem] font-black uppercase tracking-[0.05em] text-slate-500">User trực thuộc</p>
            <p class="m-0 mt-1 text-[0.78rem] font-semibold text-slate-500">Danh sách user đang thuộc tuyến đại lý của bạn.</p>
          </div>
          <span class="rounded-full bg-rose-50 px-3 py-1 text-[0.72rem] font-black text-primary">{{ managedUsers.length }} user</span>
        </div>

        <div v-if="managedUsers.length === 0" class="px-4 py-6 text-[0.82rem] font-semibold text-slate-500">
          Chưa có user trực thuộc trong tuyến đại lý này.
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full text-left text-[0.78rem]">
            <thead class="bg-slate-50 text-slate-500">
              <tr>
                <th class="px-4 py-3 font-black">ID</th>
                <th class="px-4 py-3 font-black">User</th>
                <th class="px-4 py-3 font-black">SĐT</th>
                <th class="px-4 py-3 font-black">Trạng thái</th>
                <th class="px-4 py-3 font-black">Nạp đầu</th>
                <th class="px-4 py-3 font-black">Mã GD nạp đầu</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in managedUsers" :key="item.user_id" class="border-t border-slate-100">
                <td class="px-4 py-3 font-black text-on-surface">{{ item.user_id }}</td>
                <td class="px-4 py-3 font-bold text-on-surface">{{ item.name || '—' }}</td>
                <td class="px-4 py-3 font-semibold text-slate-500">{{ item.phone || '—' }}</td>
                <td class="px-4 py-3">
                  <span class="rounded-full bg-slate-100 px-2.5 py-1 text-[0.72rem] font-black text-slate-600">
                    {{ affiliateReferralStatusLabel(item.referral_status) }}
                  </span>
                </td>
                <td class="px-4 py-3 font-bold text-on-surface">
                  {{ Number(item.first_deposit_amount || 0).toLocaleString('vi-VN') }}
                </td>
                <td class="px-4 py-3 font-mono text-[0.74rem] text-slate-500">{{ item.first_deposit_transaction_no || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div v-if="isBecomeAgencyOpen" class="fixed inset-0 z-[9999] grid place-items-center bg-black/40 px-4">
        <div class="w-full max-w-[520px] rounded-[22px] bg-white p-5 shadow-xl">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h3 class="m-0 text-[1rem] font-black text-on-surface">Trở thành đại lý</h3>
              <p class="m-0 mt-1 text-xs font-bold text-slate-500">
                Nhập mã mời của nhân viên để kích hoạt tài khoản đại lý.
              </p>
            </div>
            <button class="grid h-9 w-9 place-items-center rounded-full bg-slate-100 text-slate-600" type="button" @click="isBecomeAgencyOpen = false">
              <span class="material-symbols-outlined text-[1.1rem]">close</span>
            </button>
          </div>

          <div v-if="becomeAgencyError" class="mt-3 rounded-[14px] bg-red-50 px-4 py-3 text-xs font-bold text-red-600">
            {{ becomeAgencyError }}
          </div>

          <form class="mt-4 space-y-3" @submit.prevent="submitBecomeAgency">
            <label class="grid min-h-[56px] items-center overflow-hidden rounded-[18px] bg-surface-container-low border border-slate-100">
              <input
                v-model="staffInviteCode"
                class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none font-extrabold tracking-wide"
                type="text"
                autocomplete="off"
                placeholder="Mã mời nhân viên"
              />
            </label>
            <button
              class="min-h-12 w-full rounded-[16px] bg-primary font-black text-white disabled:opacity-60"
              type="submit"
              :disabled="becomeAgencyLoading || staffInviteCode.trim().length < 6"
            >
              {{ becomeAgencyLoading ? 'Đang xử lý...' : 'Xác nhận' }}
            </button>
          </form>
        </div>
      </div>
    </section>

    <section v-if="error" class="mx-3 rounded-[14px] bg-red-50 px-4 py-3 text-sm font-semibold text-red-600">
      {{ error }}
    </section>

    <section v-if="loading" class="mx-3 rounded-[14px] border border-slate-100 bg-white px-4 py-4 text-sm font-semibold text-slate-500">
      Đang tải dữ liệu...
    </section>

    <div class="flex flex-col gap-3 px-3">
      <div
        v-if="activeTab === 'promotion'"
        v-for="item in promotionItems"
        :key="item.id"
        class="flex overflow-hidden rounded-[18px] bg-white shadow-sm border border-slate-100 transition-all active:scale-[0.99] hover:-translate-y-0.5"
      >
        <!-- Left accent bar -->
        <div class="w-1.5 flex-shrink-0 bg-[#e8404a]" />
        <div class="flex flex-1 gap-3 px-4 py-4">
          <!-- Icon -->
          <div class="flex-shrink-0 grid h-11 w-11 place-items-center rounded-[14px] mt-0.5 bg-red-50 text-[#e8404a]">
            <span class="material-symbols-outlined text-[1.3rem]">campaign</span>
          </div>
          <!-- Content -->
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-2">
              <strong class="text-[0.9rem] font-black text-on-surface leading-snug">{{ item.title }}</strong>
            </div>
            <p class="mt-1 text-[0.75rem] leading-5 text-slate-500">{{ item.excerpt || item.content }}</p>
            <p class="mt-3 text-[0.68rem] font-bold text-slate-400">{{ item.published_at || item.created_at }}</p>
          </div>
        </div>
      </div>

      <div
        v-if="activeTab === 'promotion' && !loading && promotionItems.length === 0"
        class="rounded-[14px] border border-slate-100 bg-white px-4 py-4 text-sm font-semibold text-slate-500"
      >
        Chưa có dữ liệu khuyến mãi.
      </div>

      <RouterLink
        v-if="activeTab === 'news'"
        v-for="item in newsItems"
        :key="item.id"
        :to="`/news/${item.slug}`"
        class="rounded-[18px] border border-slate-100 bg-white p-4 shadow-sm"
      >
        <h3 class="text-[0.92rem] font-black text-on-surface">{{ item.title }}</h3>
        <p class="mt-2 text-[0.78rem] leading-5 text-slate-500">{{ item.excerpt || item.content }}</p>
        <p class="mt-3 text-[0.68rem] font-bold text-slate-400">{{ item.published_at || item.created_at }}</p>
      </RouterLink>

      <div
        v-if="activeTab === 'news' && !loading && newsItems.length === 0"
        class="rounded-[14px] border border-slate-100 bg-white px-4 py-4 text-sm font-semibold text-slate-500"
      >
        Chưa có dữ liệu tin tức.
      </div>
    </div>
  </div>
</template>
