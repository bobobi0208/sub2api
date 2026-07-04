<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page: Technical Blueprint -->
  <div v-else class="blueprint relative min-h-screen overflow-x-hidden" :class="{ 'is-dark': isDark }">
    <!-- Blueprint grid + corner registration marks -->
    <div class="pointer-events-none fixed inset-0 z-0">
      <div class="grid-lines absolute inset-0"></div>
      <div class="glow-blob absolute -right-32 top-24 h-[32rem] w-[32rem]"></div>
      <div class="glow-blob absolute -left-40 bottom-0 h-[28rem] w-[28rem]"></div>
    </div>

    <div class="relative z-10 mx-auto max-w-[1180px] px-5 sm:px-8">
      <!-- ============ Header ============ -->
      <header class="flex items-center justify-between border-b border-line py-5">
        <div class="flex items-center gap-3">
          <div class="logo-frame">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-6 w-6 object-contain" />
          </div>
          <span class="mono text-[13px] font-semibold tracking-tight text-ink">{{ siteName }}</span>
        </div>

        <div class="flex items-center gap-1.5">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="icon-btn"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="sm" />
          </a>
          <button
            @click="toggleTheme"
            class="icon-btn"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="ml-1.5 inline-flex items-center gap-2 rounded-full border border-ink bg-ink py-1.5 pl-1.5 pr-3.5 text-[12px] font-medium text-paper transition-all hover:bg-transparent hover:text-ink"
          >
            <span class="flex h-5 w-5 items-center justify-center rounded-full bg-primary-500 text-[9px] font-bold text-white">{{ userInitial }}</span>
            <span class="mono">{{ t('home.dashboard') }}</span>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="mono ml-1.5 inline-flex items-center rounded-full border border-ink bg-ink px-4 py-2 text-[12px] font-medium text-paper transition-all hover:bg-transparent hover:text-ink"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </header>

      <!-- ============ Hero ============ -->
      <section class="grid items-center gap-10 border-b border-line py-16 lg:grid-cols-[1.05fr_0.95fr] lg:py-24">
        <div class="reveal">
          <div class="mb-6 flex flex-wrap items-center gap-2.5">
            <div class="section-tag">
              <span class="tag-dot"></span>{{ t('home.pricing.badge') }}
            </div>
            <div class="sla-chip">
              <Icon name="shield" size="xs" :stroke-width="2" />
              {{ t('home.stats.uptime') }} {{ t('home.stats.uptimeLabel') }}
            </div>
          </div>
          <h1 class="display text-[2.75rem] leading-[1.02] text-ink sm:text-6xl lg:text-[4.25rem]">
            {{ t('home.heroSubtitle') }}
          </h1>
          <p class="mt-6 max-w-md text-[15px] leading-relaxed text-muted">
            {{ t('home.heroDescription') }}
          </p>

          <div class="mt-9 flex flex-wrap items-center gap-3">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="mono group inline-flex items-center gap-2.5 rounded-full bg-primary-600 px-7 py-3.5 text-[13px] font-semibold text-white shadow-lg shadow-primary-500/25 transition-all hover:bg-primary-700 hover:shadow-primary-500/35"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
              <Icon name="arrowRight" size="sm" class="transition-transform group-hover:translate-x-1" :stroke-width="2.2" />
            </router-link>
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="mono inline-flex items-center gap-2 rounded-full border border-line px-6 py-3.5 text-[13px] font-semibold text-ink transition-colors hover:border-ink"
            >
              <Icon name="book" size="sm" :stroke-width="2" /> {{ t('home.docs') }}
            </a>
          </div>

          <!-- Inline stat strip -->
          <div class="mono mt-10 flex flex-wrap gap-x-8 gap-y-3 border-t border-line pt-6 text-[11px] uppercase tracking-wider text-faint">
            <div><span class="text-primary-500">04+</span> Providers</div>
            <div><span class="text-primary-500">01</span> API Key</div>
            <div><span class="text-primary-500">∞</span> Failover</div>
          </div>
        </div>

        <!-- Terminal card -->
        <div class="reveal reveal-2 flex justify-center lg:justify-end">
          <div class="terminal">
            <div class="terminal-bar">
              <div class="dots"><span></span><span></span><span></span></div>
              <span class="mono terminal-title">POST /v1/messages</span>
            </div>
            <div class="terminal-body mono">
              <div class="ln ln-1"><span class="c-prompt">$</span> <span class="c-cmd">curl</span> <span class="c-flag">-X POST</span> <span class="c-url">/v1/messages</span></div>
              <div class="ln ln-2"><span class="c-flag">-H</span> <span class="c-str">"Authorization: Bearer sk-••••"</span></div>
              <div class="ln ln-3"><span class="c-dim"># routing → upstream pool</span></div>
              <div class="ln ln-4"><span class="c-ok">200 OK</span> <span class="c-res">{ "content": "Hello!" }</span></div>
              <div class="ln ln-5"><span class="c-prompt">$</span> <span class="caret"></span></div>
            </div>
          </div>
        </div>
      </section>

      <!-- ============ Reliability / SLA band ============ -->
      <section class="border-b border-line py-10">
        <div class="mb-6 flex items-center gap-2.5">
          <span class="live-pip"></span>
          <span class="mono text-[11px] uppercase tracking-[0.14em] text-muted">{{ t('home.stats.subtitle') }}</span>
        </div>
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <div v-for="s in stats" :key="s.key" class="stat-cell">
            <div class="mb-3 text-primary-500"><Icon :name="s.icon" size="md" :stroke-width="1.8" /></div>
            <div class="display stat-num">{{ t(`home.stats.${s.key}`) }}</div>
            <div class="mono mt-1.5 text-[11px] uppercase tracking-wider text-faint">{{ t(`home.stats.${s.key}Label`) }}</div>
          </div>
        </div>
      </section>

      <!-- ============ Pain Points ============ -->
      <section class="border-b border-line py-16">
        <div class="section-head">
          <span class="mono section-index">[ 01 ]</span>
          <h2 class="display section-title">{{ t('home.painPoints.title') }}</h2>
        </div>
        <div class="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="(p, i) in painPoints" :key="p.key" class="cell group">
            <span class="mono cell-num">{{ String(i + 1).padStart(2, '0') }}</span>
            <div class="my-4 h-8 w-8 text-muted transition-colors group-hover:text-primary-500">
              <Icon :name="p.icon" size="lg" :stroke-width="1.6" />
            </div>
            <h3 class="text-[15px] font-semibold text-ink">{{ t(`home.painPoints.items.${p.key}.title`) }}</h3>
            <p class="mt-2 text-[13px] leading-relaxed text-muted">{{ t(`home.painPoints.items.${p.key}.desc`) }}</p>
          </div>
        </div>
      </section>

      <!-- ============ Solutions / Features ============ -->
      <section class="border-b border-line py-16">
        <div class="section-head">
          <span class="mono section-index">[ 02 ]</span>
          <div>
            <h2 class="display section-title">{{ t('home.solutions.title') }}</h2>
            <p class="mt-1 text-[14px] text-muted">{{ t('home.solutions.subtitle') }}</p>
          </div>
        </div>
        <div class="mt-10 grid gap-6 md:grid-cols-3">
          <div v-for="(f, i) in features" :key="f.key" class="feature-card group">
            <div class="flex items-center justify-between">
              <div class="feat-icon">
                <Icon :name="f.icon" size="md" class="text-white" :stroke-width="1.8" />
              </div>
              <span class="mono text-[11px] text-faint">0{{ i + 1 }} / 03</span>
            </div>
            <h3 class="mt-6 text-[17px] font-semibold text-ink">{{ t(`home.features.${f.key}`) }}</h3>
            <p class="mt-2.5 text-[13.5px] leading-relaxed text-muted">{{ t(`home.features.${f.key}Desc`) }}</p>
            <div class="feat-rule"></div>
          </div>
        </div>
      </section>

      <!-- ============ Comparison ============ -->
      <section class="border-b border-line py-16">
        <div class="section-head">
          <span class="mono section-index">[ 03 ]</span>
          <h2 class="display section-title">{{ t('home.comparison.title') }}</h2>
        </div>
        <div class="mt-10 overflow-hidden rounded-2xl border border-line">
          <table class="w-full border-collapse text-left">
            <thead>
              <tr class="mono text-[11px] uppercase tracking-wider">
                <th class="cmp-th">{{ t('home.comparison.headers.feature') }}</th>
                <th class="cmp-th text-muted">{{ t('home.comparison.headers.official') }}</th>
                <th class="cmp-th cmp-th-us">{{ t('home.comparison.headers.us') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in comparisonRows" :key="row" class="cmp-row">
                <td class="cmp-td font-semibold text-ink">{{ t(`home.comparison.items.${row}.feature`) }}</td>
                <td class="cmp-td text-muted">
                  <span class="mr-2 inline-block text-faint"><Icon name="x" size="xs" :stroke-width="2.5" /></span>
                  {{ t(`home.comparison.items.${row}.official`) }}
                </td>
                <td class="cmp-td cmp-td-us text-ink">
                  <span class="mr-2 inline-block text-primary-500"><Icon name="check" size="xs" :stroke-width="2.5" /></span>
                  {{ t(`home.comparison.items.${row}.us`) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- ============ Providers ============ -->
      <section class="border-b border-line py-16">
        <div class="section-head">
          <span class="mono section-index">[ 04 ]</span>
          <div>
            <h2 class="display section-title">{{ t('home.providers.title') }}</h2>
            <p class="mt-1 text-[14px] text-muted">{{ t('home.providers.description') }}</p>
          </div>
        </div>
        <div class="mt-10 flex flex-wrap gap-3">
          <div
            v-for="p in providers"
            :key="p.label"
            class="provider-chip"
            :class="{ 'provider-soon': p.soon }"
          >
            <span class="provider-badge" :style="{ background: p.color }">{{ p.mark }}</span>
            <span class="text-[13.5px] font-semibold text-ink">{{ p.label }}</span>
            <span class="mono provider-tag" :class="p.soon ? 'provider-tag-soon' : 'provider-tag-ok'">
              {{ p.soon ? t('home.providers.soon') : t('home.providers.supported') }}
            </span>
          </div>
        </div>
      </section>

      <!-- ============ Pricing (live from admin plans; hidden when none) ============ -->
      <section v-if="plans.length" class="border-b border-line py-16">
        <div class="section-head">
          <span class="mono section-index">[ 05 ]</span>
          <div>
            <h2 class="display section-title">{{ t('home.pricing.title') }}</h2>
            <p class="mt-1.5 max-w-lg text-[14px] text-muted">{{ t('home.pricing.subtitle') }}</p>
          </div>
        </div>

        <div class="mt-10 grid gap-6" :class="planGridClass">
          <div
            v-for="(plan, i) in plans"
            :key="plan.id"
            class="price-card"
            :class="{ 'price-card-featured': i === bestValueIndex }"
          >
            <div v-if="i === bestValueIndex" class="price-ribbon">{{ t('home.pricing.popular') }}</div>

            <div class="mono text-[12px] uppercase tracking-[0.12em]" :class="i === bestValueIndex ? 'text-white/70' : 'text-faint'">
              {{ plan.name }}
            </div>

            <div class="mt-2.5 flex flex-wrap items-end gap-x-2 gap-y-1">
              <span class="display text-[2.6rem] leading-none" :class="i === bestValueIndex ? 'text-white' : 'text-ink'">
                {{ money(plan.price) }}
              </span>
              <span class="mono mb-1.5 text-[12px]" :class="i === bestValueIndex ? 'text-white/60' : 'text-faint'">
                · {{ effectiveDays(plan) }} {{ t('home.pricing.daysUnit') }}
              </span>
              <span
                v-if="plan.original_price && plan.original_price > plan.price"
                class="mono mb-1.5 text-[12px] line-through"
                :class="i === bestValueIndex ? 'text-white/45' : 'text-faint'"
              >{{ money(plan.original_price) }}</span>
            </div>

            <div class="mono mt-2 text-[11px]" :class="i === bestValueIndex ? 'text-white/65' : 'text-primary-600'">
              {{ t('home.pricing.perDay', { price: money(plan.price / effectiveDays(plan)) }) }}
            </div>

            <p v-if="plan.description" class="mt-3 text-[13px] leading-relaxed" :class="i === bestValueIndex ? 'text-white/70' : 'text-muted'">
              {{ plan.description }}
            </p>

            <div class="price-rule" :class="i === bestValueIndex ? 'bg-white/15' : 'bg-line'"></div>

            <ul class="space-y-3">
              <li
                v-for="(feat, fi) in planFeatures(plan)"
                :key="fi"
                class="flex items-start gap-2.5 text-[13.5px]"
                :class="i === bestValueIndex ? 'text-white/85' : 'text-muted'"
              >
                <span class="mt-0.5 shrink-0" :class="i === bestValueIndex ? 'text-white' : 'text-primary-500'">
                  <Icon name="check" size="xs" :stroke-width="2.5" />
                </span>
                {{ feat }}
              </li>
            </ul>

            <router-link
              :to="isAuthenticated ? dashboardPath : '/register'"
              class="mono mt-7 flex items-center justify-center gap-2 rounded-full py-3.5 text-[13px] font-semibold transition-all"
              :class="i === bestValueIndex
                ? 'bg-white text-[#0f2e2a] shadow-lg shadow-black/10 hover:-translate-y-0.5'
                : 'border border-line text-ink hover:border-primary-500 hover:text-primary-600'"
            >
              {{ t('home.pricing.cta') }}
              <Icon name="arrowRight" size="xs" :stroke-width="2.4" />
            </router-link>
          </div>
        </div>

        <p class="mono mt-8 flex items-center justify-center gap-2 text-center text-[12px] text-faint">
          <Icon name="shield" size="xs" :stroke-width="2" class="text-primary-500" />
          {{ t('home.pricing.guarantee') }}
        </p>
      </section>

      <!-- ============ CTA ============ -->
      <section class="py-16">
        <div class="cta-band">
          <div class="cta-grid"></div>
          <div class="relative z-10">
            <h2 class="display text-3xl text-white sm:text-[2.5rem]">{{ t('home.cta.title') }}</h2>
            <p class="mt-3 max-w-lg text-[14.5px] leading-relaxed text-white/70">{{ t('home.cta.description') }}</p>
          </div>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/register'"
            class="mono relative z-10 inline-flex shrink-0 items-center gap-2.5 rounded-full bg-white px-7 py-4 text-[13px] font-bold text-[#0f2e2a] shadow-lg shadow-black/15 transition-transform hover:-translate-y-0.5"
          >
            {{ isAuthenticated ? t('home.goToDashboard') : t('home.cta.button') }}
            <Icon name="arrowRight" size="sm" :stroke-width="2.4" />
          </router-link>
        </div>
      </section>

      <!-- ============ Footer ============ -->
      <footer class="flex flex-col items-center justify-between gap-4 border-t border-line py-8 sm:flex-row">
        <p class="mono text-[12px] text-faint">
          © {{ currentYear }} {{ siteName }} — {{ t('home.footer.allRightsReserved') }}
        </p>
        <div v-if="docUrl" class="mono flex items-center gap-6 text-[12px]">
          <a
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-muted transition-colors hover:text-primary-500"
          >{{ t('home.docs') }}</a>
        </div>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import type { SubscriptionPlan } from '@/types/payment'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// ---- Static content maps (labels come from i18n) ----
const painPoints = [
  { key: 'expensive', icon: 'creditCard' as const },
  { key: 'complex', icon: 'grid' as const },
  { key: 'unstable', icon: 'bolt' as const },
  { key: 'noControl', icon: 'chart' as const }
]

const features = [
  { key: 'unifiedGateway', icon: 'key' as const },
  { key: 'multiAccount', icon: 'server' as const },
  { key: 'balanceQuota', icon: 'chartBar' as const }
]

const comparisonRows = ['pricing', 'models', 'management', 'stability', 'control']

const stats = [
  { key: 'uptime', icon: 'shield' as const },
  { key: 'latency', icon: 'bolt' as const },
  { key: 'requests', icon: 'trendingUp' as const },
  { key: 'monitor', icon: 'bell' as const }
]

// ---- Subscription plans (pulled live from admin config, public endpoint) ----
const currency = '$'
const plans = ref<SubscriptionPlan[]>([])

// Convert a plan's validity into effective days for per-day price comparison.
function effectiveDays(p: SubscriptionPlan): number {
  const u = (p.validity_unit || '').toLowerCase()
  const factor = u.startsWith('week') ? 7 : u.startsWith('month') ? 30 : u.startsWith('year') ? 365 : 1
  return Math.max(1, (p.validity_days || 1) * factor)
}

// Best value = lowest per-day price; index into `plans`. -1 when 0/1 plans.
const bestValueIndex = computed(() => {
  if (plans.value.length < 2) return -1
  let best = 0
  let bestRate = plans.value[0].price / effectiveDays(plans.value[0])
  plans.value.forEach((p, i) => {
    const rate = p.price / effectiveDays(p)
    if (rate < bestRate) {
      bestRate = rate
      best = i
    }
  })
  return best
})

// Responsive grid + centering based on how many plans exist.
const planGridClass = computed(() => {
  const n = plans.value.length
  if (n <= 1) return 'mx-auto max-w-sm'
  if (n === 2) return 'mx-auto max-w-3xl sm:grid-cols-2'
  return 'sm:grid-cols-2 lg:grid-cols-3'
})

function money(n: number): string {
  const v = Math.round(n * 100) / 100
  return currency + (Number.isInteger(v) ? String(v) : String(v))
}

// backend serialises features as a JSON string; parse into a clean list.
function planFeatures(p: SubscriptionPlan): string[] {
  const raw = p.features as unknown
  if (Array.isArray(raw)) return raw.map((x) => String(x).trim()).filter(Boolean)
  if (typeof raw === 'string') {
    const s = raw.trim()
    if (!s) return []
    try {
      const parsed = JSON.parse(s)
      if (Array.isArray(parsed)) return parsed.map((x) => String(x).trim()).filter(Boolean)
    } catch {
      /* not JSON — fall through to newline split */
    }
    return s.split(/\r?\n/).map((x) => x.trim()).filter(Boolean)
  }
  return []
}

async function loadPlans() {
  try {
    const res = await paymentAPI.getPlansPublic()
    const list = (res.data || []).filter((p) => p.for_sale !== false)
    plans.value = list.slice().sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
  } catch {
    plans.value = []
  }
}

const providers = computed(() => [
  { label: t('home.providers.claude'), mark: 'C', color: 'linear-gradient(135deg,#f97316,#ea580c)', soon: false },
  { label: 'GPT', mark: 'G', color: 'linear-gradient(135deg,#10b981,#059669)', soon: false },
  { label: t('home.providers.gemini'), mark: 'G', color: 'linear-gradient(135deg,#3b82f6,#2563eb)', soon: false },
  { label: t('home.providers.antigravity'), mark: 'A', color: 'linear-gradient(135deg,#f43f5e,#db2777)', soon: false },
  { label: t('home.providers.more'), mark: '+', color: 'linear-gradient(135deg,#64748b,#475569)', soon: true }
])

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  loadPlans()
})
</script>

<style scoped>
/* ============ Design tokens ============ */
.blueprint {
  --paper: #fbfbf9;
  --ink: #101613;
  --muted: #5c665f;
  --faint: #9aa39c;
  --line: #e3e5df;
  background: var(--paper);
  font-family: 'Bricolage Grotesque', system-ui, sans-serif;
}
.blueprint.is-dark {
  --paper: #060a09;
  --ink: #f2f5f2;
  --muted: #9aa79f;
  --faint: #5a655d;
  --line: #1b211e;
  background: var(--paper);
}

.mono { font-family: 'JetBrains Mono', ui-monospace, monospace; }
.display { font-family: 'Bricolage Grotesque', system-ui, sans-serif; font-weight: 700; letter-spacing: -0.02em; }

.text-ink { color: var(--ink); }
.text-muted { color: var(--muted); }
.text-faint { color: var(--faint); }
.text-paper { color: var(--paper); }
.bg-ink { background: var(--ink); }
.border-line { border-color: var(--line); }
.bg-line { background: var(--line); }

/* ============ Background ============ */
.grid-lines {
  background-image: radial-gradient(color-mix(in srgb, var(--ink) 12%, transparent) 1.2px, transparent 1.2px);
  background-size: 30px 30px;
  -webkit-mask-image: radial-gradient(ellipse 90% 70% at 50% 0%, #000 20%, transparent 75%);
  mask-image: radial-gradient(ellipse 90% 70% at 50% 0%, #000 20%, transparent 75%);
  opacity: 0.5;
}
.glow-blob {
  border-radius: 50%;
  background: radial-gradient(circle, rgba(20, 184, 166, 0.16), transparent 68%);
  filter: blur(30px);
}

/* ============ Header bits ============ */
.logo-frame {
  display: grid;
  place-items: center;
  height: 36px;
  width: 36px;
  border-radius: 12px;
  border: 1px solid var(--line);
  background: color-mix(in srgb, var(--ink) 4%, transparent);
}
.icon-btn {
  display: grid;
  place-items: center;
  height: 36px;
  width: 36px;
  border-radius: 11px;
  color: var(--muted);
  border: 1px solid transparent;
  transition: all 0.18s ease;
}
.icon-btn:hover {
  color: var(--ink);
  border-color: var(--line);
  background: color-mix(in srgb, var(--ink) 4%, transparent);
}

/* ============ Section scaffolding ============ */
.section-tag {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.14em;
  color: var(--muted);
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 6px 14px;
}
.tag-dot {
  height: 6px;
  width: 6px;
  border-radius: 50%;
  background: #14b8a6;
  box-shadow: 0 0 0 3px rgba(20, 184, 166, 0.18);
  animation: pulse-dot 2s ease-in-out infinite;
}
@keyframes pulse-dot {
  50% { opacity: 0.35; }
}

/* SLA chip in hero */
.sla-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  letter-spacing: 0.04em;
  color: #0d9488;
  background: color-mix(in srgb, #14b8a6 12%, transparent);
  border: 1px solid color-mix(in srgb, #14b8a6 30%, transparent);
  border-radius: 999px;
  padding: 6px 13px;
}
.blueprint.is-dark .sla-chip { color: #2dd4bf; }

/* Reliability band */
.live-pip {
  position: relative;
  height: 8px;
  width: 8px;
  border-radius: 50%;
  background: #22c55e;
}
.live-pip::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: #22c55e;
  animation: ping 1.8s cubic-bezier(0, 0, 0.2, 1) infinite;
}
@keyframes ping {
  75%, 100% { transform: scale(2.4); opacity: 0; }
}
.stat-cell {
  background: var(--paper);
  border: 1px solid var(--line);
  border-radius: 20px;
  padding: 24px 22px;
  box-shadow: 0 2px 10px rgba(15, 42, 38, 0.03);
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}
.stat-cell:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, #14b8a6 40%, var(--line));
  box-shadow: 0 16px 34px -18px rgba(13, 148, 136, 0.35);
}
.stat-num {
  font-size: clamp(1.75rem, 3vw, 2.4rem);
  line-height: 1;
  color: var(--ink);
  letter-spacing: -0.02em;
}
.section-head {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}
.section-index {
  margin-top: 6px;
  font-size: 12px;
  color: #14b8a6;
  letter-spacing: 0.1em;
}
.section-title {
  font-size: clamp(1.7rem, 3.5vw, 2.5rem);
  line-height: 1.05;
  color: var(--ink);
}

/* ============ Pain point cells ============ */
.cell {
  position: relative;
  background: var(--paper);
  border: 1px solid var(--line);
  border-radius: 20px;
  padding: 26px 22px 28px;
  box-shadow: 0 2px 10px rgba(15, 42, 38, 0.03);
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}
.cell:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, #14b8a6 40%, var(--line));
  box-shadow: 0 16px 34px -18px rgba(13, 148, 136, 0.3);
}
.cell-num {
  font-size: 11px;
  color: var(--faint);
  letter-spacing: 0.1em;
}

/* ============ Feature cards ============ */
.feature-card {
  position: relative;
  border: 1px solid var(--line);
  border-radius: 24px;
  background: var(--paper);
  padding: 28px 26px 30px;
  box-shadow: 0 2px 10px rgba(15, 42, 38, 0.03);
  transition: transform 0.22s ease, border-color 0.22s ease, box-shadow 0.22s ease;
}
.feature-card:hover {
  transform: translateY(-5px);
  border-color: color-mix(in srgb, #14b8a6 45%, var(--line));
  box-shadow: 0 22px 44px -22px rgba(13, 148, 136, 0.4);
}
.feat-icon {
  display: grid;
  place-items: center;
  height: 46px;
  width: 46px;
  border-radius: 14px;
  background: linear-gradient(135deg, #14b8a6, #0d9488);
  box-shadow: 0 8px 18px -8px rgba(13, 148, 136, 0.6);
}
.feat-rule {
  margin-top: 22px;
  height: 3px;
  width: 32px;
  border-radius: 999px;
  background: #14b8a6;
  transition: width 0.28s ease;
}
.feature-card:hover .feat-rule { width: 72px; }

/* ============ Pricing cards ============ */
.price-card {
  position: relative;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--line);
  border-radius: 28px;
  background: var(--paper);
  padding: 30px 28px 32px;
  box-shadow: 0 2px 12px rgba(15, 42, 38, 0.04);
  transition: transform 0.22s ease, border-color 0.22s ease, box-shadow 0.22s ease;
}
.price-card:not(.price-card-featured):hover {
  transform: translateY(-5px);
  border-color: color-mix(in srgb, #14b8a6 45%, var(--line));
  box-shadow: 0 24px 46px -22px rgba(13, 148, 136, 0.35);
}
.price-card-featured {
  background: linear-gradient(160deg, #0f766e 0%, #0d9488 55%, #115e59 100%);
  border-color: transparent;
  box-shadow: 0 34px 64px -26px rgba(13, 148, 136, 0.6);
}
@media (min-width: 1024px) {
  .price-card-featured { transform: scale(1.04); }
  .price-card-featured:hover { transform: scale(1.04) translateY(-4px); }
}
.price-ribbon {
  position: absolute;
  top: 18px;
  right: -1px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #0f2e2a;
  background: #fff;
  padding: 6px 14px;
  border-radius: 999px 0 0 999px;
  box-shadow: 0 6px 16px -6px rgba(0, 0, 0, 0.3);
}
.price-rule {
  margin: 22px 0;
  height: 1px;
  width: 100%;
}
.price-card ul { flex: 1; }

/* ============ Comparison table ============ */
.cmp-th {
  padding: 15px 20px;
  border-bottom: 1px solid var(--line);
  color: var(--faint);
  font-weight: 500;
}
.cmp-th-us {
  color: #0d9488;
  background: color-mix(in srgb, #14b8a6 8%, transparent);
}
.blueprint.is-dark .cmp-th-us { color: #2dd4bf; }
.cmp-row { border-bottom: 1px solid var(--line); }
.cmp-row:last-child { border-bottom: 0; }
.cmp-td {
  padding: 16px 20px;
  font-size: 13.5px;
  vertical-align: middle;
}
.cmp-td-us {
  background: color-mix(in srgb, #14b8a6 6%, transparent);
  font-weight: 500;
}

/* ============ Provider chips ============ */
.provider-chip {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 9px 16px 9px 10px;
  transition: border-color 0.18s ease, transform 0.18s ease, box-shadow 0.18s ease;
}
.provider-chip:not(.provider-soon):hover {
  border-color: color-mix(in srgb, #14b8a6 45%, var(--line));
  transform: translateY(-2px);
  box-shadow: 0 12px 24px -14px rgba(13, 148, 136, 0.5);
}
.provider-soon { opacity: 0.55; }
.provider-badge {
  display: grid;
  place-items: center;
  height: 28px;
  width: 28px;
  border-radius: 10px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  font-weight: 700;
  color: #fff;
}
.provider-tag {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 3px 9px;
  border-radius: 999px;
}
.provider-tag-ok {
  color: #0d9488;
  background: color-mix(in srgb, #14b8a6 12%, transparent);
}
.blueprint.is-dark .provider-tag-ok { color: #2dd4bf; }
.provider-tag-soon {
  color: var(--faint);
  background: color-mix(in srgb, var(--ink) 6%, transparent);
}

/* ============ CTA band ============ */
.cta-band {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 28px;
  overflow: hidden;
  border-radius: 32px;
  background: linear-gradient(120deg, #0f766e 0%, #0d9488 45%, #115e59 100%);
  padding: 48px 40px;
  box-shadow: 0 34px 70px -30px rgba(13, 148, 136, 0.55);
}
@media (min-width: 768px) {
  .cta-band {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 56px 56px;
  }
}
.cta-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.08) 1px, transparent 1px);
  background-size: 40px 40px;
  -webkit-mask-image: radial-gradient(ellipse at 100% 0%, #000, transparent 70%);
  mask-image: radial-gradient(ellipse at 100% 0%, #000, transparent 70%);
}

/* ============ Terminal ============ */
.terminal {
  width: 100%;
  max-width: 440px;
  border-radius: 20px;
  border: 1px solid var(--line);
  background: #0b1120;
  overflow: hidden;
  box-shadow: 0 30px 60px -25px rgba(0, 0, 0, 0.5);
}
.blueprint.is-dark .terminal {
  box-shadow: 0 30px 60px -25px rgba(0, 0, 0, 0.8), 0 0 0 1px rgba(20, 184, 166, 0.15);
}
.terminal-bar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 11px 15px;
  background: #131c2e;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}
.dots { display: flex; gap: 7px; }
.dots span { height: 11px; width: 11px; border-radius: 50%; }
.dots span:nth-child(1) { background: #ef4444; }
.dots span:nth-child(2) { background: #eab308; }
.dots span:nth-child(3) { background: #22c55e; }
.terminal-title { font-size: 11px; color: #64748b; }
.terminal-body {
  padding: 20px 22px;
  font-size: 13px;
  line-height: 2.05;
  color: #cbd5e1;
}
.ln { opacity: 0; animation: ln-in 0.45s ease forwards; }
.ln-1 { animation-delay: 0.3s; }
.ln-2 { animation-delay: 0.9s; }
.ln-3 { animation-delay: 1.5s; }
.ln-4 { animation-delay: 2.1s; }
.ln-5 { animation-delay: 2.7s; }
@keyframes ln-in {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
.c-prompt { color: #22c55e; font-weight: 600; }
.c-cmd { color: #38bdf8; }
.c-flag { color: #a78bfa; }
.c-url { color: #2dd4bf; }
.c-str { color: #fbbf24; }
.c-dim { color: #64748b; font-style: italic; }
.c-ok {
  color: #052e16;
  background: #22c55e;
  padding: 1px 7px;
  font-weight: 700;
}
.c-res { color: #fbbf24; }
.caret {
  display: inline-block;
  width: 7px;
  height: 15px;
  background: #22c55e;
  vertical-align: middle;
  animation: blink 1s step-end infinite;
}
@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* ============ Reveal on load ============ */
.reveal {
  opacity: 0;
  transform: translateY(16px);
  animation: reveal 0.7s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
.reveal-2 { animation-delay: 0.15s; }
@keyframes reveal {
  to { opacity: 1; transform: translateY(0); }
}

@media (prefers-reduced-motion: reduce) {
  .reveal, .ln, .tag-dot, .caret { animation: none; opacity: 1; transform: none; }
}
</style>
