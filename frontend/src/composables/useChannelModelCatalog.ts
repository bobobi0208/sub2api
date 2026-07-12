import { ref } from 'vue'
import { syncPricingModels } from '@/api/admin/channels'

/**
 * 全局模型目录缓存(按平台)。
 *
 * 数据源:GET /api/v1/admin/channels/pricing/sync-models?platform=
 * (LiteLLM 定价目录 model_pricing.json)。
 *
 * 模块级缓存跨所有组件实例共享,并对并发请求做去重,避免一个渠道编辑页里
 * 多个 PricingEntryCard 同时对同一平台重复拉取。缓存仅随整页刷新失效
 * (目录变动不频繁;ChannelsView 的"同步最新模型"按钮是独立的批量刷新入口)。
 */
const catalogCache = new Map<string, string[]>()
const inflight = new Map<string, Promise<string[]>>()

function fetchCatalog(platform: string): Promise<string[]> {
  const cached = catalogCache.get(platform)
  if (cached) return Promise.resolve(cached)

  const existing = inflight.get(platform)
  if (existing) return existing

  const request = syncPricingModels(platform)
    .then((res) => {
      const models = Array.isArray(res?.models) ? res.models : []
      catalogCache.set(platform, models)
      return models
    })
    .finally(() => {
      inflight.delete(platform)
    })

  inflight.set(platform, request)
  return request
}

export function useChannelModelCatalog() {
  const models = ref<string[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  /**
   * 懒加载指定平台的模型目录(建议在下拉首次打开时调用)。
   * platform 为空时退化为不请求、目录为空(纯手动输入)。
   */
  async function ensureLoaded(platform?: string): Promise<void> {
    const key = (platform || '').trim()
    if (!key) {
      models.value = []
      error.value = null
      return
    }

    const cached = catalogCache.get(key)
    if (cached) {
      models.value = cached
      error.value = null
      return
    }

    loading.value = true
    error.value = null
    try {
      models.value = await fetchCatalog(key)
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      models.value = []
    } finally {
      loading.value = false
    }
  }

  return { models, loading, error, ensureLoaded }
}
