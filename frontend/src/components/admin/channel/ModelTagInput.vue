<template>
  <div>
    <!-- 标签展示 + 输入(点击容器聚焦输入并打开目录下拉) -->
    <div
      ref="containerRef"
      class="relative"
      @focusout="onFocusOut"
    >
      <div
        class="flex flex-wrap gap-1.5 rounded-lg border border-gray-200 bg-white p-2 dark:border-dark-600 dark:bg-dark-800 min-h-[2.5rem] cursor-text"
        @click="focusInput"
      >
        <span
          v-for="(model, idx) in models"
          :key="idx"
          class="inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-sm"
          :class="getPlatformTagClass(props.platform || '')"
        >
          {{ model }}
          <button
            type="button"
            @click.stop="removeModel(idx)"
            class="ml-0.5 rounded-full p-0.5 hover:bg-primary-200 dark:hover:bg-primary-800"
          >
            <Icon name="x" size="xs" />
          </button>
        </span>
        <input
          ref="inputRef"
          v-model="inputValue"
          type="text"
          class="flex-1 min-w-[120px] border-none bg-transparent text-sm outline-none placeholder:text-gray-400 dark:text-white"
          :placeholder="models.length === 0 ? placeholder : ''"
          @focus="openDropdown"
          @keydown.enter.prevent="addModel"
          @keydown.tab.prevent="addModel"
          @keydown.delete="handleBackspace"
          @keydown.esc="showDropdown = false"
          @paste="handlePaste"
        />
      </div>

      <!-- 目录下拉:搜索式勾选 + 自定义兜底 -->
      <div
        v-if="showDropdown"
        class="absolute left-0 right-0 top-full z-50 mt-1 max-h-64 overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-600 dark:bg-dark-700"
      >
        <div v-if="loading" class="px-3 py-4 text-center text-sm text-gray-500">
          {{ t('admin.channels.form.loadingModels', 'Loading models...') }}
        </div>
        <div v-else-if="error" class="px-3 py-4 text-center text-sm text-red-500">
          {{ t('admin.channels.form.catalogLoadError', 'Failed to load model catalog') }}
        </div>
        <template v-else>
          <!-- 目录候选(按输入过滤) -->
          <button
            v-for="model in filteredCatalog"
            :key="model"
            type="button"
            @mousedown.prevent
            @click="toggleModel(model)"
            class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-600"
          >
            <span
              :class="[
                'flex h-4 w-4 shrink-0 items-center justify-center rounded border',
                models.includes(model)
                  ? 'border-primary-500 bg-primary-500 text-white'
                  : 'border-gray-300 dark:border-dark-500'
              ]"
            >
              <Icon v-if="models.includes(model)" name="check" size="xs" />
            </span>
            <span class="truncate text-gray-900 dark:text-white">{{ model }}</span>
          </button>

          <!-- 输入了目录里没有的名字:显式提供"添加为自定义模型" -->
          <button
            v-if="showCustomRow"
            type="button"
            @mousedown.prevent
            @click="addCustomFromInput"
            class="flex w-full items-center gap-2 border-t border-gray-100 px-3 py-2 text-left text-sm text-primary-600 hover:bg-gray-100 dark:border-dark-600 dark:text-primary-400 dark:hover:bg-dark-600"
          >
            <Icon name="plus" size="xs" />
            <span class="truncate">{{ t('admin.channels.form.addAsCustomModel', { model: inputValue.trim() }) }}</span>
          </button>

          <!-- 目录非空但无匹配、且没有可添加的自定义项 -->
          <div
            v-if="filteredCatalog.length === 0 && !showCustomRow"
            class="px-3 py-4 text-center text-sm text-gray-500"
          >
            {{ catalogEmptyText }}
          </div>
        </template>
      </div>
    </div>

    <p class="mt-1 text-xs text-gray-400">
      {{ t('admin.channels.form.modelInputHint', 'Pick from the list, or type a name and press Enter to add a custom model.') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getPlatformTagClass } from './types'
import { useChannelModelCatalog } from '@/composables/useChannelModelCatalog'

const { t } = useI18n()

const props = defineProps<{
  models: string[]
  placeholder?: string
  platform?: string
}>()

const emit = defineEmits<{
  'update:models': [models: string[]]
}>()

const inputValue = ref('')
const inputRef = ref<HTMLInputElement>()
const containerRef = ref<HTMLElement>()
const showDropdown = ref(false)

const { models: catalog, loading, error, ensureLoaded } = useChannelModelCatalog()

// 目录候选:按输入做大小写不敏感子串过滤,软上限避免超长渲染
const CATALOG_RENDER_LIMIT = 200
const filteredCatalog = computed(() => {
  const query = inputValue.value.trim().toLowerCase()
  const list = query
    ? catalog.value.filter(m => m.toLowerCase().includes(query))
    : catalog.value
  return list.slice(0, CATALOG_RENDER_LIMIT)
})

// 输入了非空、目录里无精确匹配(不区分大小写)、且尚未选中 → 提供自定义添加行
const showCustomRow = computed(() => {
  const val = inputValue.value.trim()
  if (!val) return false
  const lower = val.toLowerCase()
  const inCatalog = catalog.value.some(m => m.toLowerCase() === lower)
  const alreadySelected = props.models.some(m => m.toLowerCase() === lower)
  return !inCatalog && !alreadySelected
})

const catalogEmptyText = computed(() =>
  props.platform
    ? t('admin.channels.form.noMatchingModels', 'No matching models')
    : t('admin.channels.form.noModels', 'No models added')
)

function focusInput() {
  inputRef.value?.focus()
}

function openDropdown() {
  showDropdown.value = true
  // 首次打开时懒加载该平台目录
  ensureLoaded(props.platform)
}

function onFocusOut(e: FocusEvent) {
  const next = e.relatedTarget as Node | null
  if (next && containerRef.value?.contains(next)) return
  // 焦点移出整个组件:关闭下拉,并把未提交的手打内容按原行为提交
  showDropdown.value = false
  addModel()
}

function addModel() {
  const val = inputValue.value.trim()
  if (!val) return
  if (!props.models.includes(val)) {
    emit('update:models', [...props.models, val])
  }
  inputValue.value = ''
}

// 点击"添加为自定义模型"行
function addCustomFromInput() {
  addModel()
  focusInput()
}

// 目录行勾选/取消;做出选择后清空搜索词,避免随后失焦被当成手打提交
function toggleModel(model: string) {
  if (props.models.includes(model)) {
    emit('update:models', props.models.filter(m => m !== model))
  } else {
    emit('update:models', [...props.models, model])
  }
  inputValue.value = ''
  focusInput()
}

function removeModel(idx: number) {
  const newModels = [...props.models]
  newModels.splice(idx, 1)
  emit('update:models', newModels)
}

function handleBackspace() {
  if (inputValue.value === '' && props.models.length > 0) {
    removeModel(props.models.length - 1)
  }
}

function handlePaste(e: ClipboardEvent) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text') || ''
  const items = text.split(/[,\n;]+/).map(s => s.trim()).filter(Boolean)
  if (items.length === 0) return
  const unique = [...new Set([...props.models, ...items])]
  emit('update:models', unique)
  inputValue.value = ''
}

// 平台变化时,若下拉打开则重新加载对应目录
watch(() => props.platform, (p) => {
  if (showDropdown.value) ensureLoaded(p)
})
</script>
