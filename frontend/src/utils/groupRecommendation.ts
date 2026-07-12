/**
 * 分组推荐标签(recommendation)前端共享定义。
 *
 * 与后端白名单保持一致(见 backend/internal/service/group.go):
 *   ''        无标签
 *   'featured' 主推
 *   'value'    性价比首选
 *
 * i18nKey 指向全局 messages 里的文案(见 locales 下各语言 common.ts 的 groupRecommendation),
 * 组件用 t(meta.i18nKey) 取本地化文案;badgeClass 为行内彩色 pill 的 Tailwind 类。
 */

export type GroupRecommendation = '' | 'featured' | 'value'

export interface RecommendationMeta {
  value: GroupRecommendation
  i18nKey: string
  /** 图标(emoji);无标签为空 */
  icon: string
  /** 行内实底彩色 pill 样式;'' (无标签) 为空串 */
  badgeClass: string
  /** 表单里可点选胶囊「选中态」样式(实底) */
  chipActiveClass: string
}

export const GROUP_RECOMMENDATIONS: RecommendationMeta[] = [
  { value: '', i18nKey: 'groupRecommendation.none', icon: '', badgeClass: '', chipActiveClass: 'bg-gray-600 text-white border-gray-600' },
  {
    value: 'featured',
    i18nKey: 'groupRecommendation.featured',
    icon: '⭐',
    badgeClass: 'bg-rose-500 text-white shadow-sm',
    chipActiveClass: 'bg-rose-500 text-white border-rose-500 shadow-sm'
  },
  {
    value: 'value',
    i18nKey: 'groupRecommendation.value',
    icon: '💎',
    badgeClass: 'bg-amber-500 text-white shadow-sm',
    chipActiveClass: 'bg-amber-500 text-white border-amber-500 shadow-sm'
  }
]

/** 取非空推荐标签的展示元数据;无标签或未知值返回 null。 */
export function getRecommendationMeta(
  value?: string | null
): RecommendationMeta | null {
  if (!value) return null
  return (
    GROUP_RECOMMENDATIONS.find((r) => r.value !== '' && r.value === value) ??
    null
  )
}
