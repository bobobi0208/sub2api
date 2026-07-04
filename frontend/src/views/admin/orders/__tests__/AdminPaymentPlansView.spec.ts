import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import AdminPaymentPlansView from '../AdminPaymentPlansView.vue'

const getPlans = vi.hoisted(() => vi.fn())
const getGroups = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getPlans,
    updatePlan: vi.fn(),
    deletePlan: vi.fn(),
  },
}))

vi.mock('@/api/admin', () => ({
  default: {
    groups: {
      getAll: getGroups,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess: vi.fn(),
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractI18nErrorMessage: () => 'error',
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'payment.admin.createPlan': '创建套餐',
    'payment.admin.noSubscriptionGroupsTitle': '还没有订阅类型分组',
    'payment.admin.noSubscriptionGroupsHint': '请先创建订阅类型分组，然后再创建可售卖的订阅套餐。',
    'payment.admin.createSubscriptionGroup': '创建订阅分组',
    'common.refresh': '刷新',
  }

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

const AppLayoutStub = {
  template: '<div><slot /></div>',
}

const DataTableStub = {
  props: ['columns', 'data', 'loading'],
  template: '<div><slot name="empty" /></div>',
}

const IconStub = {
  template: '<span />',
}

const RouterLinkStub = {
  props: ['to'],
  template: '<a><slot /></a>',
}

const PlanEditDialogStub = {
  props: ['show'],
  template: '<div v-if="show" data-testid="plan-edit-dialog" />',
}

function mountView() {
  return mount(AdminPaymentPlansView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        DataTable: DataTableStub,
        ConfirmDialog: true,
        Icon: IconStub,
        GroupBadge: true,
        PlanEditDialog: PlanEditDialogStub,
        RouterLink: RouterLinkStub,
      },
    },
  })
}

describe('AdminPaymentPlansView', () => {
  it('guides admins to create a subscription group before creating plans', async () => {
    getGroups.mockResolvedValue([
      {
        id: 1,
        name: 'default',
        platform: 'anthropic',
        rate_multiplier: 1,
        subscription_type: 'standard',
        status: 'active',
      },
    ])
    getPlans.mockResolvedValue({ data: [] })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('还没有订阅类型分组')
    expect(wrapper.text()).toContain('请先创建订阅类型分组')
    expect(wrapper.text()).toContain('创建订阅分组')

    const createButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('创建套餐'))

    expect(createButton?.attributes('disabled')).toBeDefined()
    await createButton?.trigger('click')
    expect(wrapper.find('[data-testid="plan-edit-dialog"]').exists()).toBe(false)
  })
})
