<script setup lang="ts">
import { computed, ref, reactive, watch } from 'vue'
import {
  NButton, NSwitch, NInput, NFlex, NModal, NCard, NTag,
  NScrollbar, NDivider, NEmpty, NSpace, NTooltip, NIcon,
  useMessage
} from 'naive-ui'
import { Add, TrashOutline, GlobeOutline, LockClosedOutline, EyeOutline, EyeOffOutline } from '@vicons/ionicons5'
import EmbeddedUrl from './EmbeddedUrl.vue'

// ── 类型定义 ───────────────────────────────────────────────
interface HotmapTab {
  name: string
  url: string
  builtin: boolean
  hidden: boolean
}

interface UserConfig {
  hiddenBuiltins: string[]          // 被隐藏的内置栏目名称列表
  customTabs: { name: string; url: string }[]  // 用户自定义栏目
}

// ── 内置栏目（不可删除，仅可隐藏）──────────────────────────
const BUILTIN_TABS: Omit<HotmapTab, 'hidden'>[] = [
  { name: '选股通',       url: 'https://xuangutong.com.cn',                              builtin: true },
  { name: '百度股市通',   url: 'https://gushitong.baidu.com',                            builtin: true },
  { name: '东财大盘星图', url: 'https://quote.eastmoney.com/stockhotmap/',               builtin: true },
  { name: 'TopHub',       url: 'https://tophub.today/c/finance',                         builtin: true },
  { name: '摸鱼',         url: 'https://996.ninja/',                                     builtin: true },
  { name: '财联社-行情',  url: 'https://www.cls.cn/quotation',                           builtin: true },
  { name: '消息墙',       url: 'https://go-stock.sparkmemory.top:16667/go-stock',        builtin: true },
  { name: '现货黄金',     url: 'https://www.tradinghero.com/chart?symbol=XAUUSD.GOODS', builtin: true },
  { name: '金十期货',     url: 'https://qihuo.jin10.com/',                               builtin: true },
  { name: '金十数据',     url: 'https://datas.jin10.com/#/category/52005',               builtin: true },
]

const STORAGE_KEY = 'go_stock_hotmap_user_config'

// ── 读取 / 初始化用户配置 ──────────────────────────────────
function loadConfig(): UserConfig {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw) as UserConfig
  } catch {}
  return { hiddenBuiltins: [], customTabs: [] }
}

const message = useMessage()
const userConfig = reactive<UserConfig>(loadConfig())

function saveConfig() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(userConfig))
}

// ── 计算最终展示的标签页 ───────────────────────────────────
const displayTabs = computed<HotmapTab[]>(() => {
  const builtins = BUILTIN_TABS
    .filter(t => !userConfig.hiddenBuiltins.includes(t.name))
    .map(t => ({ ...t, hidden: false }))
  const customs = userConfig.customTabs.map(t => ({ ...t, builtin: false, hidden: false }))
  return [...builtins, ...customs]
})

// ── 管理面板状态 ───────────────────────────────────────────
const showManage = ref(false)
const newTabName = ref('')
const newTabUrl  = ref('')
const nameError  = ref('')
const urlError   = ref('')

function isBuiltinHidden(name: string) {
  return userConfig.hiddenBuiltins.includes(name)
}

function toggleBuiltin(name: string) {
  const idx = userConfig.hiddenBuiltins.indexOf(name)
  if (idx === -1) {
    userConfig.hiddenBuiltins.push(name)
  } else {
    userConfig.hiddenBuiltins.splice(idx, 1)
  }
  saveConfig()
}

function addCustomTab() {
  nameError.value = ''
  urlError.value  = ''
  const name = newTabName.value.trim()
  const url  = newTabUrl.value.trim()

  if (!name) { nameError.value = '请输入栏目名称'; return }
  if (name.length > 20) { nameError.value = '名称不超过 20 个字符'; return }

  // 检查名称重复
  const allNames = [...BUILTIN_TABS.map(t => t.name), ...userConfig.customTabs.map(t => t.name)]
  if (allNames.includes(name)) { nameError.value = '该名称已存在'; return }

  if (!url) { urlError.value = '请输入 URL'; return }
  if (!/^https?:\/\/.+/.test(url)) { urlError.value = '请输入合法的 URL（以 http:// 或 https:// 开头）'; return }

  userConfig.customTabs.push({ name, url })
  saveConfig()
  newTabName.value = ''
  newTabUrl.value  = ''
  message.success(`已添加栏目「${name}」`)
}

function removeCustomTab(name: string) {
  const idx = userConfig.customTabs.findIndex(t => t.name === name)
  if (idx !== -1) {
    userConfig.customTabs.splice(idx, 1)
    saveConfig()
    message.success(`已删除栏目「${name}」`)
  }
}

function resetToDefault() {
  userConfig.hiddenBuiltins.splice(0)
  userConfig.customTabs.splice(0)
  saveConfig()
  message.success('已恢复默认设置')
}

const activeTab = ref(displayTabs.value[0]?.name ?? '')

// 若当前激活的 tab 被隐藏了，切换到第一个可见 tab
watch(displayTabs, (tabs) => {
  if (!tabs.find(t => t.name === activeTab.value)) {
    activeTab.value = tabs[0]?.name ?? ''
  }
})
</script>

<template>
  <!-- 动态 Tabs，管理按钮放在 suffix 插槽 -->
  <n-tabs v-model:value="activeTab" type="line" animated>
    <template #suffix>
      <n-button size="tiny" secondary type="primary" style="margin-right: 4px;" @click="showManage = true">
        <template #icon><n-icon :component="GlobeOutline" /></template>
        管理栏目
      </n-button>
    </template>
      <n-tab-pane
        v-for="tab in displayTabs"
        :key="tab.name"
        :name="tab.name"
        :tab="tab.name"
      >
        <embedded-url :url="tab.url" :height="'calc(100vh - 252px)'" />
      </n-tab-pane>
      <n-tab-pane v-if="displayTabs.length === 0" name="__empty__" tab="暂无栏目">
        <n-empty description="所有内置栏目已隐藏，且没有自定义栏目" style="margin-top: 80px;">
          <template #extra>
            <n-button @click="showManage = true" type="primary">去管理栏目</n-button>
          </template>
        </n-empty>
      </n-tab-pane>
  </n-tabs>

  <!-- 管理面板 Modal -->
  <n-modal
    v-model:show="showManage"
    preset="card"
    title="管理名站优选栏目"
    style="width: 560px; max-width: calc(100vw - 32px);"
    :mask-closable="true"
  >
    <n-scrollbar style="max-height: 65vh;">
      <!-- 内置栏目 -->
      <div style="margin-bottom: 4px;">
        <n-flex align="center" :size="6" style="margin-bottom: 8px;">
          <n-icon :component="LockClosedOutline" style="color: #999;" />
          <span style="font-size: 13px; font-weight: 600; color: #666;">内置栏目</span>
          <n-tag size="tiny" :bordered="false" type="info">不可删除，可隐藏</n-tag>
        </n-flex>

        <div
          v-for="tab in BUILTIN_TABS"
          :key="tab.name"
          style="display: flex; align-items: center; justify-content: space-between; padding: 6px 4px; border-radius: 6px;"
          :style="isBuiltinHidden(tab.name) ? 'opacity: 0.45;' : ''"
        >
          <n-flex align="center" :size="8">
            <n-switch
              :value="!isBuiltinHidden(tab.name)"
              size="small"
              @update:value="() => toggleBuiltin(tab.name)"
            />
            <span style="font-size: 13px;">{{ tab.name }}</span>
            <n-tag size="tiny" :bordered="false" style="color: #aaa; max-width: 220px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" :title="tab.url">
              {{ tab.url }}
            </n-tag>
          </n-flex>
          <n-tooltip :delay="600">
            <template #trigger>
              <n-icon
                :component="isBuiltinHidden(tab.name) ? EyeOffOutline : EyeOutline"
                style="cursor: pointer; color: #999; font-size: 16px;"
                @click="toggleBuiltin(tab.name)"
              />
            </template>
            {{ isBuiltinHidden(tab.name) ? '点击显示' : '点击隐藏' }}
          </n-tooltip>
        </div>
      </div>

      <n-divider style="margin: 12px 0;" />

      <!-- 用户自定义栏目 -->
      <n-flex align="center" :size="6" style="margin-bottom: 8px;">
        <n-icon :component="Add" style="color: #2080f0;" />
        <span style="font-size: 13px; font-weight: 600; color: #666;">自定义栏目</span>
        <n-tag size="tiny" :bordered="false" type="success">{{ userConfig.customTabs.length }} 个</n-tag>
      </n-flex>

      <n-empty
        v-if="userConfig.customTabs.length === 0"
        description="暂无自定义栏目，在下方添加"
        size="small"
        style="margin: 8px 0 12px 0;"
      />

      <div
        v-for="tab in userConfig.customTabs"
        :key="tab.name"
        style="display: flex; align-items: center; justify-content: space-between; padding: 6px 4px; border-radius: 6px;"
      >
        <n-flex align="center" :size="8" style="overflow: hidden;">
          <span style="font-size: 13px; white-space: nowrap;">{{ tab.name }}</span>
          <n-tag size="tiny" :bordered="false" style="color: #aaa; max-width: 220px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" :title="tab.url">
            {{ tab.url }}
          </n-tag>
        </n-flex>
        <n-tooltip :delay="600">
          <template #trigger>
            <n-icon
              :component="TrashOutline"
              style="cursor: pointer; color: #d03050; font-size: 16px; flex-shrink: 0;"
              @click="removeCustomTab(tab.name)"
            />
          </template>
          删除该栏目
        </n-tooltip>
      </div>

      <n-divider style="margin: 12px 0 8px 0;">添加新栏目</n-divider>

      <!-- 添加表单 -->
      <n-space vertical :size="6">
        <div>
          <n-input
            v-model:value="newTabName"
            placeholder="栏目名称（最多 20 字）"
            clearable
            size="small"
            :status="nameError ? 'error' : undefined"
            @keydown.enter="addCustomTab"
          />
          <div v-if="nameError" style="font-size: 11px; color: #d03050; margin-top: 2px;">{{ nameError }}</div>
        </div>
        <div>
          <n-input
            v-model:value="newTabUrl"
            placeholder="URL（例：https://example.com）"
            clearable
            size="small"
            :status="urlError ? 'error' : undefined"
            @keydown.enter="addCustomTab"
          />
          <div v-if="urlError" style="font-size: 11px; color: #d03050; margin-top: 2px;">{{ urlError }}</div>
        </div>
        <n-button type="primary" size="small" block @click="addCustomTab">
          <template #icon><n-icon :component="Add" /></template>
          确认添加
        </n-button>
      </n-space>
    </n-scrollbar>

    <template #footer>
      <n-flex justify="space-between" align="center">
        <n-button size="small" type="error" secondary @click="resetToDefault">
          恢复默认设置
        </n-button>
        <n-button size="small" @click="showManage = false">关闭</n-button>
      </n-flex>
    </template>
  </n-modal>
</template>

<style scoped>
</style>