import { test, expect } from '@playwright/test'

// See here how to get started:
// https://playwright.dev/docs/intro
test('visits the app root url', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('h1')).toHaveText('You did it!')
})

const eventTokenKey = 'fh88u:wheel-event-token'

function eventState(overrides: Record<string, unknown> = {}) {
  const now = Date.now()

  return {
    server_now: new Date(now).toISOString(),
    invitation_id: 'invitation-e2e',
    campaign_name: 'Sự kiện E2E',
    session_id: 'session-e2e',
    session_status: 'active',
    started_at: new Date(now - 10_000).toISOString(),
    ends_at: new Date(now + 240_000).toISOString(),
    current_round: 2,
    next_round_available_at: new Date(now - 500).toISOString(),
    spin_duration_seconds: 5,
    rounds: [
      {
        round_no: 1,
        status: 'spun',
        segment_key: 'reward_500k',
        result_label: '500.000 đồng',
        prize_amount: '500000',
        spun_at: new Date(now - 6_000).toISOString(),
      },
      { round_no: 2, status: 'pending' },
      { round_no: 3, status: 'pending' },
    ],
    paid_rewards: [{ round_no: 1, amount: '500000', status: 'paid' }],
    total_reward: '500000',
    ...overrides,
  }
}

async function seedEventToken(page: import('@playwright/test').Page) {
  await page.addInitScript(
    ({ key }) => {
      if (window.location.pathname.endsWith('/event.html')) {
        window.sessionStorage.setItem(key, 'e2e-wheel-token')
      }
    },
    { key: eventTokenKey },
  )
}

test('launch transition page renders without JavaScript', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 812 })
  await page.goto('/event-launching.html')

  await expect(page.getByRole('heading', { name: 'Đang chuẩn bị vé tham gia...' })).toBeVisible()
  for (const alt of ['Bình giữ nhiệt logo pp789i', 'AirPods Pro', 'Xe VinFast Limo Green']) {
    await expect(page.getByRole('img', { name: alt })).toBeVisible()
  }
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
    true,
  )
})

test('launch failure stays visible instead of closing the tab', async ({ page }) => {
  await page.goto('/event-launching.html?failed=1')

  await expect(page.getByRole('heading', { name: 'Không thể mở sự kiện' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Đóng và thử lại' })).toBeVisible()
})

test('event html has an immediate boot preview before the app mounts', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 812 })
  await page.route(/\/assets\/event-[^/]+\.js$/, (route) => route.abort())

  await page.goto('/event.html')

  await expect(page.locator('#event-boot')).toBeVisible()
  await expect(page.getByText('Đang xác thực vé tham gia của bạn')).toBeVisible()
  await expect(page.locator('#event-boot img')).toHaveCount(4)
})

test('event loading previews the real prize assets', async ({ page }, testInfo) => {
  await page.setViewportSize({ width: 375, height: 812 })
  await seedEventToken(page)
  let releaseState!: () => void
  const stateGate = new Promise<void>((resolve) => {
    releaseState = resolve
  })
  await page.route('**/v1/wheel/me', async (route) => {
    await stateGate
    await route.fulfill({ json: eventState() })
  })
  await page.route('**/v1/wheel/session/chat/messages**', (route) =>
    route.fulfill({ json: { items: [] } }),
  )
  await page.route('**/v1/wheel/realtime/ticket', (route) =>
    route.fulfill({ status: 503, json: { message: 'offline' } }),
  )

  await page.goto('/event.html')

  await expect(page.getByText('Đang mở phòng sự kiện...')).toBeVisible()
  for (const alt of ['Bình giữ nhiệt logo pp789i', 'AirPods Pro', 'Xe VinFast Limo Green']) {
    const image = page.locator(`.event-loading__prizes img[alt="${alt}"]`)
    await expect(image).toBeVisible()
    await expect
      .poll(() => image.evaluate((element: HTMLImageElement) => element.naturalWidth))
      .toBeGreaterThan(0)
    expect(
      await image.evaluate((element: HTMLImageElement) => {
        const canvas = document.createElement('canvas')
        canvas.width = element.naturalWidth
        canvas.height = element.naturalHeight
        const context = canvas.getContext('2d')
        context?.drawImage(element, 0, 0)
        return context?.getImageData(0, 0, 1, 1).data[3]
      }),
    ).toBe(0)
  }
  await page.screenshot({ path: testInfo.outputPath('event-loading-preview.png'), fullPage: true })

  releaseState()
  await expect(page.getByText('Vòng quay may mắn')).toBeVisible()
  await expect(page.getByText('39 TRIỆU')).toBeVisible()
  await expect(page.getByText('ĐẶC BIỆT · LIMO')).toBeVisible()
  const giftImages = page.locator('.wheel-label__img')
  await expect(giftImages).toHaveCount(3)
  for (let index = 0; index < 3; index += 1) {
    expect((await giftImages.nth(index).boundingBox())?.width ?? 0).toBeGreaterThanOrEqual(45.9)
  }
  await page.screenshot({ path: testInfo.outputPath('event-wheel-mobile.png'), fullPage: true })
})

test('event chat stays bounded and scrolls with many long messages', async ({ page }, testInfo) => {
  await page.setViewportSize({ width: 1440, height: 900 })
  await seedEventToken(page)
  await page.route('**/v1/wheel/me', (route) => route.fulfill({ json: eventState() }))
  await page.route('**/v1/wheel/session/chat/messages**', (route) =>
    route.fulfill({
      json: {
        items: Array.from({ length: 60 }, (_, index) => ({
          id: index + 1,
          display_name: `ID game #${120000 + index}`,
          body: `Tin nhắn ${index + 1}: ${'nội dung dài cần xuống dòng '.repeat(12)}`,
          actor_type: 'bot',
          created_at: new Date().toISOString(),
        })),
      },
    }),
  )
  await page.route('**/v1/wheel/realtime/ticket', (route) =>
    route.fulfill({ status: 503, json: { message: 'offline' } }),
  )

  await page.goto('/event.html')
  await expect(page.getByText('Tin nhắn 60:', { exact: false })).toBeVisible()

  const chat = page.locator('.event-chat')
  const messages = page.locator('.event-chat__messages')
  const composer = page.locator('.event-chat__composer')
  const chatBox = await chat.boundingBox()
  const composerBox = await composer.boundingBox()
  const scrollState = await messages.evaluate((element) => ({
    clientHeight: element.clientHeight,
    scrollHeight: element.scrollHeight,
  }))

  expect(chatBox?.height ?? 0).toBeLessThanOrEqual(781)
  expect(scrollState.scrollHeight).toBeGreaterThan(scrollState.clientHeight)
  expect((composerBox?.y ?? 0) + (composerBox?.height ?? 0)).toBeLessThanOrEqual(
    (chatBox?.y ?? 0) + (chatBox?.height ?? 0) + 1,
  )
  expect(await page.evaluate(() => document.documentElement.scrollHeight)).toBeLessThan(1800)
  await page.screenshot({
    path: testInfo.outputPath('event-chat-scroll-desktop.png'),
    fullPage: true,
  })
})

test('stale pending round becomes clickable without reloading', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 812 })
  await seedEventToken(page)
  let stateCalls = 0
  await page.route('**/v1/wheel/me', (route) => route.fulfill({ json: eventState() }))
  await page.route('**/v1/wheel/session/state', (route) => {
    stateCalls += 1
    if (stateCalls === 1) return route.fulfill({ status: 503, json: { message: 'temporary' } })

    return route.fulfill({
      json: eventState({
        rounds: [
          {
            round_no: 1,
            status: 'spun',
            segment_key: 'reward_500k',
            result_label: '500.000 đồng',
            prize_amount: '500000',
            spun_at: new Date(Date.now() - 6_000).toISOString(),
          },
          { round_no: 2, status: 'ready' },
          { round_no: 3, status: 'pending' },
        ],
      }),
    })
  })
  await page.route('**/v1/wheel/session/chat/messages**', (route) =>
    route.fulfill({ json: { items: [] } }),
  )
  await page.route('**/v1/wheel/realtime/ticket', (route) =>
    route.fulfill({ status: 503, json: { message: 'offline' } }),
  )

  await page.goto('/event.html')

  const spinButton = page.getByRole('button', { name: 'Quay lượt 2' })
  await expect(spinButton).toBeVisible()
  await expect(spinButton).toBeEnabled()
  await expect(page.getByText('Đang đồng bộ...')).toHaveCount(0)
  await expect.poll(() => stateCalls).toBeGreaterThanOrEqual(2)
})

test('completed event only shows the terminal screen', async ({ page }, testInfo) => {
  await page.setViewportSize({ width: 375, height: 812 })
  await seedEventToken(page)
  const completed = eventState({
    session_status: 'completed',
    current_round: 3,
    next_round_available_at: undefined,
    rounds: [1, 2, 3].map((round) => ({
      round_no: round,
      status: 'spun',
      segment_key: 'reward_500k',
      result_label: '500.000 đồng',
      prize_amount: '500000',
      spun_at: new Date().toISOString(),
    })),
  })
  await page.route('**/v1/wheel/me', (route) => route.fulfill({ json: completed }))

  await page.goto('/event.html')

  await expect(page.getByRole('heading', { name: 'KẾT THÚC' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Về trang chủ' })).toBeVisible()
  await expect(page.getByText('Vòng quay may mắn')).toHaveCount(0)
  await expect(page.getByText('Phòng trò chuyện')).toHaveCount(0)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
    true,
  )
  await page.screenshot({ path: testInfo.outputPath('event-finished-mobile.png'), fullPage: true })
  await page.getByRole('button', { name: 'Về trang chủ' }).click()
  await expect(page).toHaveURL(/\/home$/)
  await expect
    .poll(() => page.evaluate((key) => window.sessionStorage.getItem(key), eventTokenKey))
    .toBeNull()
})

test('the third reward leads to the locked terminal screen', async ({ page }) => {
  test.setTimeout(20_000)
  await page.setViewportSize({ width: 375, height: 812 })
  await seedEventToken(page)
  const active = eventState({
    current_round: 3,
    next_round_available_at: undefined,
    rounds: [
      {
        round_no: 1,
        status: 'spun',
        segment_key: 'try_again',
        result_label: 'Chúc bạn may mắn',
        prize_amount: '0',
        spun_at: new Date(Date.now() - 16_000).toISOString(),
      },
      {
        round_no: 2,
        status: 'spun',
        segment_key: 'reward_39m',
        result_label: '39 triệu đồng',
        prize_amount: '39000000',
        spun_at: new Date(Date.now() - 10_000).toISOString(),
      },
      { round_no: 3, status: 'ready' },
    ],
  })
  const completedRound = {
    round_no: 3,
    status: 'spun',
    segment_key: 'reward_68k',
    result_label: '68.000 đồng',
    prize_amount: '68000',
    spun_at: new Date().toISOString(),
  }
  const completed = eventState({
    session_status: 'completed',
    current_round: 3,
    next_round_available_at: undefined,
    rounds: [...active.rounds.slice(0, 2), completedRound],
    paid_rewards: [
      { round_no: 2, amount: '39000000', status: 'paid' },
      { round_no: 3, amount: '68000', status: 'paid' },
    ],
    total_reward: '39068000',
  })
  await page.route('**/v1/wheel/me', (route) => route.fulfill({ json: active }))
  await page.route('**/v1/wheel/session/rounds/3/spin', (route) =>
    route.fulfill({ json: { state: completed, result: completedRound } }),
  )
  await page.route('**/v1/wheel/session/chat/messages**', (route) =>
    route.fulfill({ json: { items: [] } }),
  )
  await page.route('**/v1/wheel/realtime/ticket', (route) =>
    route.fulfill({ status: 503, json: { message: 'offline' } }),
  )

  await page.goto('/event.html')
  await page.getByRole('button', { name: 'Quay lượt 3' }).click()
  await expect(page.getByRole('heading', { name: 'Bạn vừa trúng thưởng!' })).toBeVisible({
    timeout: 7_000,
  })
  await page.getByRole('button', { name: 'Kết thúc' }).click()

  await expect(page.getByRole('heading', { name: 'KẾT THÚC' })).toBeVisible()
  await expect(page.getByText('Phòng trò chuyện')).toHaveCount(0)
})

test('event chat hides actor names and shows game IDs', async ({ page }) => {
  await seedEventToken(page)
  await page.route('**/v1/wheel/me', (route) =>
    route.fulfill({
      json: eventState({
        session_id: undefined,
        session_status: 'pending',
        current_round: 1,
        next_round_available_at: undefined,
      }),
    }),
  )
  await page.route('**/v1/wheel/session/chat/messages**', (route) =>
    route.fulfill({
      json: {
        items: [
          {
            id: 1,
            display_name: 'Tám Lì',
            body: 'Chờ vòng quay nào',
            actor_type: 'bot',
            created_at: new Date().toISOString(),
          },
        ],
      },
    }),
  )
  await page.route('**/v1/wheel/realtime/ticket', (route) =>
    route.fulfill({ status: 503, json: { message: 'offline' } }),
  )

  await page.goto('/event.html')

  await expect(page.getByText(/^ID game #[0-9]{6}$/)).toBeVisible()
  await expect(page.getByText('Tám Lì')).toHaveCount(0)
})
