# Home Subscription I18n Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复首页翻译键直出，将订阅套餐前移，并把 Hero 改为订阅制卖点。

**Architecture:** 保持现有 Vue 组件、公共套餐 API 和套餐卡样式不变。先用语言包契约测试和 DOM 顺序测试复现，再补齐 `home.stats/home.pricing` 双语文案并移动现有模板区块。

**Tech Stack:** Vue 3、Vue I18n、Vitest、Vue Test Utils、TypeScript、Vite

## Global Constraints

- 中文主标题必须为“一次订阅，畅享专属 AI 额度”。
- 套餐区必须位于 Hero 之后、运行状态区之前。
- 套餐数据源、显隐条件、购买入口、价格计算与样式保持不变。
- 中英文必须同时补齐，禁止回退到直接显示翻译键。
- 生产部署必须单独取得明确确认。

---

### Task 1: 锁定首页语言包契约

**Files:**
- Create: `frontend/src/i18n/__tests__/homeLocaleKeys.spec.ts`
- Modify: `frontend/src/i18n/locales/zh/landing.ts`
- Modify: `frontend/src/i18n/locales/en/landing.ts`

**Interfaces:**
- Consumes: `zhLanding.home`、`enLanding.home` 语言包对象。
- Produces: 完整的 `home.stats`、`home.pricing` 和订阅制 Hero 文案。

- [ ] **Step 1: 写缺键和 Hero 文案的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import enLanding from '../locales/en/landing'
import zhLanding from '../locales/zh/landing'

const requiredPaths = [
  'stats.subtitle', 'stats.uptime', 'stats.uptimeLabel',
  'stats.latency', 'stats.latencyLabel', 'stats.requests',
  'stats.requestsLabel', 'stats.monitor', 'stats.monitorLabel',
  'pricing.badge', 'pricing.title', 'pricing.subtitle',
  'pricing.popular', 'pricing.cta', 'pricing.perDay',
  'pricing.daysUnit', 'pricing.guarantee',
]

function getPath(root: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((value, key) =>
    value && typeof value === 'object'
      ? (value as Record<string, unknown>)[key]
      : undefined, root)
}

describe.each([
  ['zh', zhLanding.home],
  ['en', enLanding.home],
])('home locale %s', (_, home) => {
  it.each(requiredPaths)('defines %s', (path) => {
    expect(getPath(home, path)).toEqual(expect.any(String))
    expect(getPath(home, path)).not.toBe('')
  })
})

it('uses the approved subscription hero copy', () => {
  expect(zhLanding.home.heroSubtitle).toBe('一次订阅，畅享专属 AI 额度')
  expect(enLanding.home.heroSubtitle).toBe('One Subscription, Dedicated AI Capacity')
})
```

- [ ] **Step 2: 运行测试并确认因缺键/旧文案失败**

Run: `pnpm test:run src/i18n/__tests__/homeLocaleKeys.spec.ts`

Expected: FAIL，指出 `stats.subtitle` 为 `undefined`，Hero 仍是旧文案。

- [ ] **Step 3: 补齐中英文语言包**

在两个 `landing.ts` 的 `home.providers` 后添加 `stats` 与 `pricing`，并更新：

```ts
// zh
heroSubtitle: '一次订阅，畅享专属 AI 额度',
heroDescription: '日卡、周卡、月卡灵活选择，订阅期内享专属额度，到期自动失效且不自动续费',

// en
heroSubtitle: 'One Subscription, Dedicated AI Capacity',
heroDescription: 'Choose a daily, weekly, or monthly plan with dedicated quota, automatic expiry, and no auto-renewal',
```

`stats/pricing` 使用以下完整对象：

```ts
// zh
stats: {
  subtitle: '所有系统运行正常 · 7×24 小时监控',
  uptime: '99.99%', uptimeLabel: '服务可用性 SLA',
  latency: '<200ms', latencyLabel: '中位延迟',
  requests: '千万级', requestsLabel: '每日请求',
  monitor: '24/7', monitorLabel: '实时监控',
},
pricing: {
  badge: '灵活订阅',
  title: '选择适合你的订阅套餐',
  subtitle: '日卡、周卡、月卡灵活选择 · 专属额度 · 到期自动失效，不自动续费',
  popular: '最超值', cta: '立即订阅',
  perDay: '折合 {price} / 天', daysUnit: '天',
  guarantee: '购买即时开通 · 支持支付宝 / 微信 / Stripe · 到期不自动续费',
},

// en
stats: {
  subtitle: 'All systems operational · monitored 24/7',
  uptime: '99.99%', uptimeLabel: 'Service availability SLA',
  latency: '<200ms', latencyLabel: 'Median latency',
  requests: '10M+', requestsLabel: 'Daily requests',
  monitor: '24/7', monitorLabel: 'Live monitoring',
},
pricing: {
  badge: 'Flexible Subscriptions',
  title: 'Choose the Subscription That Fits',
  subtitle: 'Daily, weekly, and monthly plans · dedicated quota · automatic expiry with no auto-renewal',
  popular: 'Best Value', cta: 'Subscribe Now',
  perDay: '{price} / day', daysUnit: 'days',
  guarantee: 'Instant activation · Alipay / WeChat Pay / Stripe · no auto-renewal',
},
```

- [ ] **Step 4: 运行语言包测试并确认通过**

Run: `pnpm test:run src/i18n/__tests__/homeLocaleKeys.spec.ts src/i18n/__tests__/localesNoKeyCollision.spec.ts`

Expected: PASS。

### Task 2: 前移套餐区并固化顺序

**Files:**
- Modify: `frontend/src/views/__tests__/HomeView.spec.ts`
- Modify: `frontend/src/views/HomeView.vue`

**Interfaces:**
- Consumes: `plans.length`、现有套餐卡模板和 `home.pricing.*` 翻译键。
- Produces: `Hero -> Pricing -> Reliability -> Pain Points -> Solutions -> Comparison -> Providers -> CTA` 顺序。

- [ ] **Step 1: 写 DOM 顺序失败测试**

先把现有套餐对象提取为：

```ts
const subscriptionPlanFixture = {
  id: 1,
  group_id: 7,
  group_platform: 'openai',
  name: 'GPT Daily',
  description: 'Daily plan',
  price: 1.91,
  original_price: 1.91,
  validity_days: 1,
  validity_unit: 'days',
  features: [],
  for_sale: true,
  sort_order: 1,
}
```

复用该 fixture，在价格测试后追加：

```ts
it('renders subscription plans before the reliability band', async () => {
  getPlansPublic.mockResolvedValue({ data: [subscriptionPlanFixture] })
  const wrapper = mountHomeView()
  await flushPromises()

  const content = wrapper.text()
  expect(content.indexOf('home.pricing.title')).toBeGreaterThan(-1)
  expect(content.indexOf('home.stats.subtitle')).toBeGreaterThan(-1)
  expect(content.indexOf('home.pricing.title')).toBeLessThan(
    content.indexOf('home.stats.subtitle'),
  )
})
```

- [ ] **Step 2: 运行测试并确认当前顺序失败**

Run: `pnpm test:run src/views/__tests__/HomeView.spec.ts`

Expected: FAIL，`home.pricing.title` 的索引大于 `home.stats.subtitle`。

- [ ] **Step 3: 移动现有套餐模板并重排编号**

把完整 `v-if="plans.length"` 套餐 `<section>` 移到 Hero 结束标签之后、Reliability 注释之前。套餐编号改为 `[ 01 ]`，痛点、方案、对比、服务商依次改为 `[ 02 ]`、`[ 03 ]`、`[ 04 ]`、`[ 05 ]`。

- [ ] **Step 4: 运行首页测试并确认通过**

Run: `pnpm test:run src/views/__tests__/HomeView.spec.ts`

Expected: PASS。

### Task 3: 全量验证与页面检查

**Files:**
- Verify only: `frontend/`

**Interfaces:**
- Consumes: Task 1-2 的最终前端代码。
- Produces: 可发布的前端构建与桌面/移动截图证据。

- [ ] **Step 1: 运行完整前端测试与静态检查**

Run: `pnpm test:run && pnpm typecheck && pnpm lint:check`

Expected: 所有命令退出码 0。

- [ ] **Step 2: 运行生产构建**

Run: `pnpm build`

Expected: `vue-tsc -b && vite build` 退出码 0。

- [ ] **Step 3: 启动预览并检查桌面/移动布局**

Run: `pnpm preview --host 127.0.0.1 --port 4173`

检查 1440×900 与 390×844：无 `home.stats.*`/`home.pricing.*` 直出，套餐在状态区之前，主标题正确，无文字溢出和区块重叠。

- [ ] **Step 4: 提交实现**

```bash
git add frontend/src/i18n/__tests__/homeLocaleKeys.spec.ts \
  frontend/src/i18n/locales/zh/landing.ts \
  frontend/src/i18n/locales/en/landing.ts \
  frontend/src/views/__tests__/HomeView.spec.ts \
  frontend/src/views/HomeView.vue
git commit -m "fix(frontend): repair home subscription translations"
```
