import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import HomeView from '../HomeView.vue'

const getPlansPublic = vi.hoisted(() => vi.fn())
const checkAuth = vi.hoisted(() => vi.fn())
const fetchPublicSettings = vi.hoisted(() => vi.fn())

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getPlansPublic,
  },
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    isAuthenticated: false,
    isAdmin: false,
    user: null,
    checkAuth,
  }),
  useAppStore: () => ({
    cachedPublicSettings: null,
    siteName: 'Sub2API',
    siteLogo: '',
    docUrl: '',
    publicSettingsLoaded: true,
    fetchPublicSettings,
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) =>
        key === 'home.pricing.perDay' ? `per day ${params?.price}` : key,
    }),
  }
})

const RouterLinkStub = {
  props: ['to'],
  template: '<a><slot /></a>',
}

const LocaleSwitcherStub = {
  template: '<span />',
}

const IconStub = {
  template: '<span />',
}

function mountHomeView() {
  return mount(HomeView, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        LocaleSwitcher: LocaleSwitcherStub,
        Icon: IconStub,
      },
    },
  })
}

describe('HomeView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn().mockReturnValue({
        matches: false,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      }),
    })
  })

  it('renders public subscription plan prices in USD on the landing page', async () => {
    getPlansPublic.mockResolvedValue({
      data: [
        {
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
        },
      ],
    })

    const wrapper = mountHomeView()
    await flushPromises()

    expect(wrapper.text()).toContain('$1.91')
    expect(wrapper.text()).not.toContain('¥1.91')
  })
})
