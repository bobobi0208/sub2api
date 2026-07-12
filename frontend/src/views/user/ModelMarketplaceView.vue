<template>
  <AppLayout>
    <div class="model-marketplace min-h-full px-4 py-6 sm:px-6 lg:px-8">
      <div class="mx-auto max-w-[1600px] space-y-6">
        <section class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.28em] text-emerald-600/80 dark:text-emerald-300/80">
              {{ t('modelMarketplace.kicker') }}
            </p>
            <h1 class="mt-2 text-2xl font-black tracking-tight text-slate-950 dark:text-white sm:text-3xl">
              {{ t('modelMarketplace.title') }}
            </h1>
            <p class="mt-2 max-w-2xl text-sm text-slate-600 dark:text-slate-300">
              {{ t('modelMarketplace.description') }}
            </p>
          </div>

          <button
            type="button"
            class="inline-flex items-center justify-center gap-2 rounded-2xl border border-white/80 bg-white/90 px-5 py-3 text-sm font-bold text-slate-700 shadow-lg shadow-slate-200/70 transition hover:-translate-y-0.5 hover:bg-white dark:border-white/10 dark:bg-dark-800/90 dark:text-slate-100 dark:shadow-black/20"
            :disabled="loading"
            @click="loadChannels"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            {{ t('modelMarketplace.refresh') }}
          </button>
        </section>

        <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div class="market-stat-card">
            <div class="market-stat-icon bg-violet-100 text-violet-600 dark:bg-violet-500/15 dark:text-violet-300">
              <Icon name="cube" size="lg" />
            </div>
            <div>
              <p class="market-stat-label">{{ t('modelMarketplace.totalModels') }}</p>
              <p class="market-stat-value">{{ marketplace.totalModels }}</p>
            </div>
          </div>

          <div
            v-for="stat in topPlatformStats"
            :key="stat.platform"
            class="market-stat-card"
          >
            <div class="market-stat-icon" :class="platformIconShellClass(stat.platform)">
              <PlatformIcon :platform="stat.platform as GroupPlatform" size="lg" />
            </div>
            <div>
              <p class="market-stat-label uppercase">{{ platformLabel(stat.platform) }}</p>
              <p class="market-stat-value">{{ stat.count }}</p>
            </div>
          </div>
        </section>

        <section class="rounded-[2rem] border border-white/80 bg-white/90 p-4 shadow-xl shadow-slate-200/70 backdrop-blur dark:border-white/10 dark:bg-dark-900/80 dark:shadow-black/20">
          <div class="grid gap-4 xl:grid-cols-[1fr_260px_auto] xl:items-center">
            <label class="relative block">
              <Icon
                name="search"
                size="lg"
                class="pointer-events-none absolute left-5 top-1/2 -translate-y-1/2 text-slate-400"
              />
              <input
                v-model="searchQuery"
                type="search"
                class="h-14 w-full rounded-2xl border border-slate-200 bg-white pl-14 pr-4 text-base font-semibold text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-emerald-300 focus:ring-4 focus:ring-emerald-100 dark:border-dark-600 dark:bg-dark-800 dark:text-white dark:focus:ring-emerald-500/10"
                :placeholder="t('modelMarketplace.searchPlaceholder')"
              />
            </label>

            <label class="relative block">
              <Icon
                name="cube"
                size="md"
                class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
              />
              <select
                v-model="selectedGroupValue"
                class="h-14 w-full appearance-none rounded-2xl border border-slate-200 bg-white pl-12 pr-10 text-base font-bold text-slate-800 outline-none transition focus:border-emerald-300 focus:ring-4 focus:ring-emerald-100 dark:border-dark-600 dark:bg-dark-800 dark:text-white dark:focus:ring-emerald-500/10"
              >
                <option value="all">{{ t('modelMarketplace.allGroups') }}</option>
                <option
                  v-for="group in marketplace.groups"
                  :key="group.id"
                  :value="String(group.id)"
                >
                  {{ group.name }}
                </option>
              </select>
              <Icon
                name="chevronDown"
                size="md"
                class="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-slate-400"
              />
            </label>

            <label class="flex h-14 items-center justify-between gap-4 rounded-2xl border border-slate-200 bg-white px-4 text-sm font-bold text-slate-700 dark:border-dark-600 dark:bg-dark-800 dark:text-slate-200">
              <span>{{ t('modelMarketplace.showRatePrice') }}</span>
              <input v-model="showRatePrice" type="checkbox" class="peer sr-only" />
              <span class="relative h-7 w-12 rounded-full bg-slate-200 transition peer-checked:bg-emerald-500 dark:bg-dark-600">
                <span
                  class="absolute left-1 top-1 h-5 w-5 rounded-full bg-white shadow transition"
                  :class="showRatePrice ? 'translate-x-5' : ''"
                ></span>
              </span>
            </label>
          </div>
        </section>

        <section class="flex flex-wrap gap-3">
          <button
            type="button"
            class="market-platform-tab"
            :class="selectedPlatform === 'all' ? 'market-platform-tab-active' : ''"
            @click="selectedPlatform = 'all'"
          >
            <Icon name="grid" size="md" />
            {{ t('modelMarketplace.allPlatforms') }}
            <span>{{ filteredByGroup.length }}</span>
          </button>
          <button
            v-for="stat in marketplace.platforms"
            :key="stat.platform"
            type="button"
            class="market-platform-tab"
            :class="selectedPlatform === stat.platform ? 'market-platform-tab-active' : ''"
            @click="selectedPlatform = stat.platform"
          >
            <PlatformIcon :platform="stat.platform as GroupPlatform" size="md" />
            {{ platformLabel(stat.platform) }}
            <span>{{ stat.count }}</span>
          </button>
        </section>

        <section v-if="loading" class="grid gap-5 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
          <div v-for="idx in 8" :key="idx" class="market-skeleton-card" />
        </section>

        <section
          v-else-if="visibleCards.length > 0"
          class="grid gap-5 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4"
        >
          <article
            v-for="card in visibleCards"
            :key="card.id"
            class="market-model-card"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex min-w-0 items-start gap-3">
                <div class="market-model-icon" :class="platformIconShellClass(card.platform)">
                  <PlatformIcon :platform="card.platform as GroupPlatform" size="lg" />
                </div>
                <div class="min-w-0">
                  <h2 class="truncate text-lg font-black text-slate-950 dark:text-white" :title="card.name">
                    {{ card.name }}
                  </h2>
                  <p class="mt-0.5 text-xs font-bold uppercase tracking-[0.22em] text-slate-400">
                    {{ platformLabel(card.platform) }}
                  </p>
                </div>
              </div>

              <div
                v-if="rateInfo(card)"
                class="shrink-0 rounded-full border border-violet-200 bg-violet-50 px-3 py-1 text-xs font-black text-violet-600 dark:border-violet-500/30 dark:bg-violet-500/10 dark:text-violet-300"
                :title="rateInfo(card)?.label"
              >
                {{ t('modelMarketplace.rateBadge', { rate: formatRate(rateInfo(card)?.rate ?? 1) }) }}
              </div>
            </div>

            <div class="mt-5 flex flex-wrap gap-2">
              <span v-if="card.hasPromptCaching" class="market-feature-pill">
                <Icon name="sparkles" size="xs" />
                {{ t('modelMarketplace.promptCaching') }}
              </span>
              <span class="market-channel-pill" :title="card.channels.join(', ')">
                {{ card.channels[0] }}
              </span>
            </div>

            <div class="mt-6 space-y-4">
              <PriceLine
                :label="t('modelMarketplace.input')"
                icon="upload"
                color-class="text-emerald-500"
                :value="formatPrice(card, 'input_price')"
                :unit="priceUnit(card)"
              />
              <PriceLine
                :label="t('modelMarketplace.output')"
                icon="download"
                color-class="text-blue-500"
                :value="formatPrice(card, 'output_price')"
                :unit="priceUnit(card)"
              />

              <div class="border-t border-dashed border-slate-200 pt-4 dark:border-dark-600">
                <PriceLine
                  :label="t('modelMarketplace.cacheWrite')"
                  icon="document"
                  color-class="text-orange-400"
                  :value="formatPrice(card, 'cache_write_price')"
                  :unit="priceUnit(card)"
                  muted
                />
                <PriceLine
                  class="mt-3"
                  :label="t('modelMarketplace.cacheRead')"
                  icon="cloud"
                  color-class="text-violet-400"
                  :value="formatPrice(card, 'cache_read_price')"
                  :unit="priceUnit(card)"
                  muted
                />
              </div>
            </div>
          </article>
        </section>

        <section
          v-else
          class="rounded-[2rem] border border-dashed border-slate-300 bg-white/70 p-12 text-center shadow-lg shadow-slate-200/50 dark:border-dark-600 dark:bg-dark-900/60 dark:shadow-black/20"
        >
          <Icon name="inbox" size="xl" class="mx-auto text-slate-400" />
          <p class="mt-4 text-sm font-semibold text-slate-500 dark:text-slate-300">
            {{ t('modelMarketplace.empty') }}
          </p>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import userChannelsAPI, { type UserAvailableChannel } from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { BILLING_MODE_IMAGE, BILLING_MODE_PER_REQUEST, BILLING_MODE_TOKEN } from '@/constants/channel'
import type { GroupPlatform } from '@/types'
import {
  buildModelMarketplace,
  getCardEffectiveRate,
  getFeaturedPlatformStats,
  getMarketplacePrice,
  type MarketplaceGroupFilter,
  type MarketplaceModelCard,
  type MarketplacePriceKey,
} from '@/utils/modelMarketplace'
import { platformLabel } from '@/utils/platformColors'

const PriceLine = defineComponent({
  props: {
    label: { type: String, required: true },
    icon: { type: String, required: true },
    colorClass: { type: String, required: true },
    value: { type: String, required: true },
    unit: { type: String, required: true },
    muted: { type: Boolean, default: false },
  },
  setup(props, { attrs }) {
    return () =>
      h('div', { class: ['flex items-center justify-between gap-3', attrs.class] }, [
        h('div', { class: ['flex items-center gap-2 text-sm font-semibold', props.muted ? 'text-slate-400' : 'text-slate-500 dark:text-slate-300'] }, [
          h(Icon, { name: props.icon as any, size: 'md', class: props.colorClass }),
          h('span', props.label),
        ]),
        h('div', { class: ['font-mono text-lg font-black', props.muted ? 'text-slate-500 dark:text-slate-300' : 'text-slate-950 dark:text-white'] }, [
          props.value,
          h('span', { class: 'ml-0.5 text-xs font-bold text-slate-400' }, props.unit),
        ]),
      ])
  },
})

const { t } = useI18n()
const appStore = useAppStore()

const channels = ref<UserAvailableChannel[]>([])
const userGroupRates = ref<Record<number, number>>({})
const loading = ref(false)
const searchQuery = ref('')
const selectedPlatform = ref('all')
const selectedGroupValue = ref('all')
const showRatePrice = ref(false)

const selectedGroupId = computed<MarketplaceGroupFilter>(() => {
  if (selectedGroupValue.value === 'all') return 'all'
  return Number(selectedGroupValue.value)
})

const marketplace = computed(() => buildModelMarketplace(channels.value, userGroupRates.value))

const topPlatformStats = computed(() => getFeaturedPlatformStats(marketplace.value.platforms))

const filteredByGroup = computed(() => {
  if (selectedGroupId.value === 'all') return marketplace.value.cards
  return marketplace.value.cards.filter((card) =>
    card.groups.some((group) => group.id === selectedGroupId.value),
  )
})

const visibleCards = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return filteredByGroup.value.filter((card) => {
    if (selectedPlatform.value !== 'all' && card.platform !== selectedPlatform.value) return false
    if (!q) return true
    return (
      card.name.toLowerCase().includes(q) ||
      card.platform.toLowerCase().includes(q) ||
      platformLabel(card.platform).toLowerCase().includes(q) ||
      card.channels.some((channel) => channel.toLowerCase().includes(q)) ||
      card.groups.some((group) => group.name.toLowerCase().includes(q))
    )
  })
})

async function loadChannels() {
  loading.value = true
  try {
    const [list, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch((err: unknown) => {
        console.error('Failed to load user group rates:', err)
        return {} as Record<number, number>
      }),
    ])
    channels.value = list
    userGroupRates.value = rates
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

function rateInfo(card: MarketplaceModelCard) {
  return getCardEffectiveRate(card, selectedGroupId.value)
}

function formatRate(rate: number): string {
  return Number(rate.toPrecision(4)).toString()
}

function formatPrice(card: MarketplaceModelCard, key: MarketplacePriceKey): string {
  const value = getMarketplacePrice(card, key, selectedGroupId.value, showRatePrice.value)
  if (value == null) return '-'

  const mode = card.pricing?.billing_mode
  if (mode === BILLING_MODE_TOKEN) {
    const perMillion = value * 1_000_000
    const digits = perMillion > 0 && perMillion < 1 ? 3 : 2
    return `$${perMillion.toFixed(digits)}`
  }

  const digits = value > 0 && value < 1 ? 4 : 2
  return `$${value.toFixed(digits)}`
}

function priceUnit(card: MarketplaceModelCard): string {
  switch (card.pricing?.billing_mode) {
    case BILLING_MODE_TOKEN:
      return t('modelMarketplace.unitMillion')
    case BILLING_MODE_PER_REQUEST:
      return t('modelMarketplace.unitRequest')
    case BILLING_MODE_IMAGE:
      return t('modelMarketplace.unitImage')
    default:
      return ''
  }
}

function platformIconShellClass(platform: string): string {
  switch (platform) {
    case 'anthropic':
      return 'bg-orange-50 text-orange-500 dark:bg-orange-500/15 dark:text-orange-300'
    case 'openai':
      return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'gemini':
      return 'bg-blue-50 text-blue-600 dark:bg-blue-500/15 dark:text-blue-300'
    case 'antigravity':
      return 'bg-violet-50 text-violet-600 dark:bg-violet-500/15 dark:text-violet-300'
    default:
      return 'bg-slate-100 text-slate-600 dark:bg-slate-500/15 dark:text-slate-300'
  }
}

onMounted(loadChannels)
</script>

<style scoped>
.model-marketplace {
  background:
    radial-gradient(circle at 18% 10%, rgba(45, 212, 191, 0.2), transparent 28rem),
    radial-gradient(circle at 85% 0%, rgba(125, 211, 252, 0.2), transparent 26rem),
    linear-gradient(135deg, rgba(236, 253, 245, 0.95), rgba(248, 250, 252, 0.96) 45%, rgba(241, 245, 249, 0.98));
}

:global(.dark) .model-marketplace {
  background:
    radial-gradient(circle at 20% 0%, rgba(20, 184, 166, 0.16), transparent 26rem),
    radial-gradient(circle at 80% 0%, rgba(59, 130, 246, 0.14), transparent 26rem),
    linear-gradient(135deg, rgba(2, 6, 23, 0.98), rgba(15, 23, 42, 0.98));
}

.market-stat-card {
  @apply flex items-center gap-5 rounded-[1.75rem] border border-white/80 bg-white/90 p-5 shadow-xl shadow-slate-200/70 backdrop-blur transition hover:-translate-y-0.5 dark:border-white/10 dark:bg-dark-900/80 dark:shadow-black/20;
}

.market-stat-icon,
.market-model-icon {
  @apply flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl;
}

.market-stat-label {
  @apply text-sm font-extrabold text-slate-500 dark:text-slate-300;
}

.market-stat-value {
  @apply mt-1 text-4xl font-black leading-none text-slate-950 dark:text-white;
}

.market-platform-tab {
  @apply inline-flex items-center gap-2 rounded-2xl border border-white/80 bg-white/90 px-5 py-3 text-sm font-black text-slate-600 shadow-lg shadow-slate-200/60 transition hover:-translate-y-0.5 hover:text-slate-950 dark:border-white/10 dark:bg-dark-900/80 dark:text-slate-300 dark:shadow-black/20 dark:hover:text-white;
}

.market-platform-tab span {
  @apply rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-500 dark:bg-dark-700 dark:text-slate-300;
}

.market-platform-tab-active {
  @apply border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-300;
}

.market-model-card {
  @apply rounded-[1.75rem] border border-white/90 bg-white/95 p-5 shadow-xl shadow-slate-200/70 backdrop-blur transition duration-200 hover:-translate-y-1 hover:shadow-2xl hover:shadow-slate-300/60 dark:border-white/10 dark:bg-dark-900/85 dark:shadow-black/20 dark:hover:shadow-black/30;
}

.market-feature-pill {
  @apply inline-flex items-center gap-1 rounded-lg bg-amber-50 px-3 py-1.5 text-xs font-black text-amber-600 dark:bg-amber-500/10 dark:text-amber-300;
}

.market-channel-pill {
  @apply max-w-[12rem] truncate rounded-lg bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-500 dark:bg-dark-700 dark:text-slate-300;
}

.market-skeleton-card {
  @apply h-80 animate-pulse rounded-[1.75rem] border border-white/70 bg-white/70 shadow-xl shadow-slate-200/60 dark:border-white/10 dark:bg-dark-900/60 dark:shadow-black/20;
}
</style>
