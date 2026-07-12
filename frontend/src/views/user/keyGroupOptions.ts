/**
 * 把「分组下拉选项」按分类(category)分段,用于 API 密钥创建的分组下拉。
 *
 * 复用 Select.vue 原生的分段头能力:带 `kind:'group'` 且 `disabled:true` 的 option
 * 会被渲染成不可选的 section header(见 apiKeyGroupFilterOptions.ts 同款用法)。
 *
 * 约定:
 * - 传入的 options 已按后端 sort_order 排好序;分类首次出现的顺序即分类段的顺序
 *   (等价于「该分类内最小 sort_order」),因此复用现成的拖拽排序即可控制段序。
 * - 未分类(category 为空)统一归到末尾的「未分类」段。
 * - 若不存在任何非空分类,则不加任何段头,退化为原来的平铺列表(平滑兼容)。
 * - 段头用递减的负数哨兵值,避免与真实分组 id 冲突。
 */

export type KeyGroupSectionHeader = {
  value: number
  label: string
  kind: 'group'
  disabled: true
}

export function buildSectionedGroupOptions<T extends { category?: string }>(
  options: T[],
  uncategorizedLabel: string
): Array<T | KeyGroupSectionHeader> {
  // 按分类分桶,保持首次出现顺序
  const order: string[] = []
  const buckets = new Map<string, T[]>()
  const uncategorized: T[] = []

  for (const opt of options) {
    const cat = (opt.category || '').trim()
    if (!cat) {
      uncategorized.push(opt)
      continue
    }
    if (!buckets.has(cat)) {
      buckets.set(cat, [])
      order.push(cat)
    }
    buckets.get(cat)!.push(opt)
  }

  // 没有任何分类 → 平铺,不加段头(与旧行为一致)
  if (order.length === 0) return options

  const result: Array<T | KeyGroupSectionHeader> = []
  let sentinel = -1
  for (const cat of order) {
    result.push({ value: sentinel--, label: cat, kind: 'group', disabled: true })
    result.push(...buckets.get(cat)!)
  }
  if (uncategorized.length > 0) {
    result.push({
      value: sentinel--,
      label: uncategorizedLabel,
      kind: 'group',
      disabled: true
    })
    result.push(...uncategorized)
  }
  return result
}
