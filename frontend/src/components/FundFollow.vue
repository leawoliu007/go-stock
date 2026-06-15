<script setup>
import {computed, h, nextTick, onBeforeMount, onBeforeUnmount, onMounted, reactive, ref} from "vue";
import {Add, ChatboxOutline, RefreshOutline} from "@vicons/ionicons5";
import {
  NAvatar,
  NButton,
  NFlex,
  NForm,
  NFormItem,
  NGradientText,
  NInput,
  NInputNumber,
  NModal,
  NSpin,
  NSwitch,
  NSelect,
  NText,
  useDialog,
  useMessage,
  useNotification
} from 'naive-ui';
import {
  Environment,
  EventsEmit,
  EventsOff,
  EventsOn,
  WindowFullscreen,
  WindowReload,
  WindowUnfullscreen
} from "../../wailsjs/runtime";
import {
  AddGroup,
  FollowFund,
  GetAiConfigs,
  GetAIResponseResult,
  GetConfig,
  GetFollowedFund,
  GetFundHistoryNetValue,
  GetFundTop10Holdings,
  GetfundList,
  GetPromptTemplates,
  GetVersionInfo,
  NewChatStream,
  OpenURL,
  SaveAIResponseResult,
  SaveAsMarkdown,
  SaveImage,
  SaveWordFile,
  ShareAnalysis,
  UnFollowFund
} from "../../wailsjs/go/main/App";
import {MdEditor, MdPreview} from 'md-editor-v3';
import {ExportPDF} from '@vavt/v3-extension';
import '@vavt/v3-extension/lib/asset/ExportPDF.css';
import html2canvas from "html2canvas";
import {asBlob} from 'html-docx-js-typescript';
import vueDanmaku from 'vue3-danmaku'
import FundKlineChart from "./FundKlineChart.vue";

const danmus = ref([])
const ws = ref(null)
const icon = ref(null)
const dialog = useDialog()
const notify = useNotification()
const message = useMessage()
const chartModalShow = ref(false)
const chartFundCode = ref('')
const chartFundName = ref('')
const netValueData = ref([])
const netValueLoading = ref(false)
const darkTheme = ref(false)
const showPopover = ref(false)
const holdingsMap = reactive({})
const modalShow4 = ref(false)
const toolbars = [0]
const handleProgress = (progress) => {}
const enableEditor = ref(false)
const mdPreviewRef = ref(null)
const mdEditorRef = ref(null)
const aiResultScrollRef = ref(null)
const tipsRef = ref(null)
const enableTools = ref(true)
const thinkingMode = ref(true)
const promptTemplates = ref([])
const aiConfigs = ref([])
const sysPromptOptions = ref([])
const userPromptOptions = ref([])

const data = reactive({
  modelName: "",
  chatId: "",
  question: "",
  sysPromptId: null,
  aiConfigId: 0,
  name: "",
  code: "",
  fullscreen: false,
  airesult: "",
  openAiEnable: false,
  loading: true,
  enableDanmu: false,
  changePercent: 0,
  time: ""
})

const followList = ref([])
const followLoading = ref(false)
const options = ref([])
const ticker = ref({})
const REFRESH_INTERVAL = 60
const countdown = ref(REFRESH_INTERVAL)
const refreshing = ref(false)
const countdownTimer = ref({})

function loadFollowedFunds() {
  followLoading.value = true
  GetFollowedFund().then(result => {
    followList.value = result || []
    loadCurrentPageHoldings()
  }).finally(() => {
    followLoading.value = false
  })
}

function loadCurrentPageHoldings() {
  for (const fund of followList.value) {
    if (!holdingsMap[fund.code]) {
      loadHoldings(fund.code)
    }
  }
}

const netValueColumns = computed(() => {
  const onExchange = isOnExchangeFund(chartFundCode.value)
  const cols = [
    { title: '日期', key: 'date', width: 110 },
    { title: onExchange ? '收盘价' : '单位净值', key: 'netValue', width: 100, render: (row) => row.netValue ?? '-' },
  ]
  if (!onExchange) {
    cols.push({ title: '累计净值', key: 'accumValue', width: 100, render: (row) => row.accumValue ?? '-' })
  }
  cols.push({
    title: '日涨幅',
    key: 'dailyGrowth',
    width: 90,
    render: (row) => {
      const v = row.dailyGrowth
      if (v == null) return '-'
      const color = v > 0 ? '#ef5350' : v < 0 ? '#26a69a' : undefined
      return h(NText, { style: { color } }, () => v.toFixed(2) + '%')
    }
  })
  if (!onExchange) {
    cols.push({ title: '申购', key: 'buyStatus', width: 70 })
    cols.push({ title: '赎回', key: 'sellStatus', width: 70 })
  }
  return cols
})

onBeforeMount(() => {
  GetConfig().then(result => {
    if (result.openAiEnable) data.openAiEnable = true
    if (result.enableDanmu) data.enableDanmu = true
    if (result.darkTheme) darkTheme.value = true
  })
  GetPromptTemplates("", "").then(res => {
    promptTemplates.value = res
    sysPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型系统 Prompt')
    userPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型用户 Prompt')
  }).catch(err => { console.error("GetPromptTemplates error:", err) })
  GetAiConfigs().then(res => {
    aiConfigs.value = res
    if (res && res.length > 0) {
      data.aiConfigId = res[0].ID
    }
  }).catch(err => { console.error("GetAiConfigs error:", err) })
  loadFollowedFunds()
})

onMounted(() => {
  GetVersionInfo().then((res) => {
    icon.value = res.icon
  })

  ws.value = new WebSocket('ws://8.134.249.145:16688/ws');
  ws.value.onopen = () => {}
  ws.value.onmessage = (event) => {
    if (data.enableDanmu) danmus.value.push(event.data)
  }
  ws.value.onerror = (error) => { console.error('WebSocket 错误:', error) }
  ws.value.onclose = () => {}

  ticker.value = setInterval(() => {
    refreshAllFunds()
  }, 1000 * REFRESH_INTERVAL)

  countdownTimer.value = setInterval(() => {
    if (countdown.value > 0) {
      countdown.value--
    }
  }, 1000)

  EventsOn("newChatStream", async (msg) => {
    if (msg === "DONE") {
      SaveAIResponseResult(data.code, data.name, data.airesult, data.chatId, data.question, data.aiConfigId)
      message.info("AI 分析完成！")
      message.destroyAll()
      data.loading = false
    } else {
      if (msg.chatId) {
        data.chatId = msg.chatId
      }
      if (msg.question) {
        data.question = msg.question
      }
      if (msg.content || msg.reasoning_content || msg.extraContent) {
        data.loading = false
      }
      if (msg.content && typeof msg.content === 'string') {
        data.airesult = data.airesult + msg.content;
      }
      if (msg.reasoning_content && typeof msg.reasoning_content === 'string') {
        data.airesult = data.airesult + msg.reasoning_content;
      }
      if (msg.extraContent && typeof msg.extraContent === 'string') {
        data.airesult = data.airesult + msg.extraContent;
      }
      nextTick(() => {
        if (aiResultScrollRef.value) {
          aiResultScrollRef.value.scrollTop = aiResultScrollRef.value.scrollHeight
        }
      })
    }
  })
})

onBeforeUnmount(() => {
  clearInterval(ticker.value)
  clearInterval(countdownTimer.value)
  if (ws.value) ws.value.close()
  message.destroyAll()
  notify.destroyAll()
  EventsOff("newChatStream")
})

function refreshAllFunds() {
  refreshing.value = true
  countdown.value = REFRESH_INTERVAL
  loadFollowedFunds()
  setTimeout(() => { refreshing.value = false }, 500)
}

function manualRefresh() {
  if (refreshing.value) return
  refreshAllFunds()
}

function preloadHoldings() {
  loadCurrentPageHoldings()
}

function SendDanmu() {
  ws.value.send(data.name)
}

function AddFund() {
  if (!data.code) {
    showPopover.value = true
    setTimeout(() => { showPopover.value = false }, 3000)
    return
  }
  FollowFund(data.code).then(result => {
    if (result) {
      message.success("关注成功")
      loadFollowedFunds()
    }
  })
}

function unFollow(code) {
  UnFollowFund(code).then(result => {
    if (result) {
      message.success("取消关注成功")
      loadFollowedFunds()
    }
  })
}

function getFundList(value) {
  GetfundList(value).then(result => {
    options.value = []
    result.forEach(item => {
      options.value.push({
        label: item.name + " [" + item.code + "]",
        value: item.code,
      })
    })
  })
}

function onSelectFund(value) {
  data.code = value
  blinkBorder(value)
}

function search(code) {
  setTimeout(() => {
    Environment().then(env => {
      switch (env.platform) {
        case 'windows':
          window.open("https://fund.eastmoney.com/" + code + ".html", "_blank", "noreferrer,width=1000,top=100,left=100,status=no,toolbar=no,location=no,scrollbars=no")
          break
        default:
          OpenURL("https://fund.eastmoney.com/" + code + ".html")
      }
    })
  }, 300)
}

function showChart(code, name) {
  chartFundCode.value = code
  chartFundName.value = name
  chartModalShow.value = true
  loadNetValueHistory(code)
}

function loadHoldings(code) {
  if (holdingsMap[code]) return
  GetFundTop10Holdings(code).then(result => {
    holdingsMap[code] = result || []
  }).catch(() => {
    holdingsMap[code] = []
  })
}

function loadNetValueHistory(code) {
  netValueLoading.value = true
  netValueData.value = []
  GetFundHistoryNetValue(code, 30, '', '').then(result => {
    netValueData.value = result || []
  }).catch(() => {
    netValueData.value = []
  }).finally(() => {
    netValueLoading.value = false
  })
}

function isOnExchangeFund(code) {
  const p = code?.substring(0, 2)
  return ['15', '16', '50', '51', '52'].includes(p)
}

function rateType(rate) {
  if (rate > 0) return 'error'
  if (rate < 0) return 'success'
  return 'default'
}

function growthType(val) {
  if (val > 0) return 'error'
  if (val < 0) return 'success'
  return 'default'
}

function ratioColor(ratio) {
  if (ratio >= 8) return '#ef5350'
  if (ratio >= 5) return '#e6a23c'
  return undefined
}

function splitHalves(arr) {
  if (!arr || !arr.length) return [[], []]
  const mid = Math.ceil(arr.length / 2)
  return [arr.slice(0, mid), arr.slice(mid)]
}

function changeText(rate) {
  if (rate == null) return '-'
  return (rate > 0 ? '+' : '') + rate.toFixed(2) + '%'
}

function changeColor(rate) {
  if (rate == null) return undefined
  return rate > 0 ? '#ef5350' : rate < 0 ? '#19b860' : undefined
}

function blinkBorder(findId) {
  const element = document.getElementById(findId)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth' })
    const pelement = document.getElementById(findId + '_gi')
    if (pelement) {
      pelement.classList.add('blink-border')
      setTimeout(() => { pelement.classList.remove('blink-border') }, 1000 * 5)
    }
  }
}

function aiReCheckFund(fund, fundCode) {
  data.modelName = ""
  data.airesult = ""
  data.time = ""
  data.name = fund
  data.code = fundCode
  data.loading = true
  modalShow4.value = true
  message.loading("ai 检测中...", {
    duration: 0,
  })
  NewChatStream(fund, fundCode, data.question, data.aiConfigId, data.sysPromptId, enableTools.value, thinkingMode.value)
}

function aiCheckFund(fund, fundCode) {
  GetAIResponseResult(fundCode).then(result => {
    if (result.content) {
      data.modelName = result.modelName
      data.chatId = result.chatId
      data.question = result.question
      data.name = fund
      data.code = fundCode
      data.loading = false
      modalShow4.value = true
      data.airesult = result.content
      const date = new Date(result.CreatedAt);
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');
      data.time = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
    } else {
      data.modelName = ""
      data.question = ""
      data.airesult = ""
      data.time = ""
      data.name = fund
      data.code = fundCode
      data.loading = false
      modalShow4.value = true
    }
  })
}

function scrollToAiResultBottom() {
  nextTick(() => {
    requestAnimationFrame(() => {
      const el = aiResultScrollRef.value
      if (el) {
        el.scrollTop = el.scrollHeight
      }
    })
  })
}

function saveAsImage(name, code) {
  const previewEl = mdPreviewRef.value?.$el || mdEditorRef.value?.$el
  const element = previewEl?.querySelector('.md-editor-preview-wrapper') ||
                  previewEl?.querySelector('.md-editor-preview') ||
                  document.querySelector('.md-editor-preview')
  if (!element) {
    message.error('无法找到分析结果元素')
    return
  }
  const savedStyles = []
  let el = element.parentElement
  while (el && el !== document.body) {
    const style = getComputedStyle(el)
    if (style.overflow === 'hidden' || style.overflowY === 'hidden' || style.overflowY === 'auto' || style.overflowY === 'scroll') {
      savedStyles.push({ el, overflow: el.style.overflow, overflowY: el.style.overflowY, height: el.style.height, maxHeight: el.style.maxHeight })
      el.style.overflow = 'visible'
      el.style.overflowY = 'visible'
      el.style.height = 'auto'
      el.style.maxHeight = 'none'
    }
    el = el.parentElement
  }
  const savedTargetStyle = { height: element.style.height, maxHeight: element.style.maxHeight, overflow: element.style.overflow, overflowY: element.style.overflowY }
  element.style.height = 'auto'
  element.style.maxHeight = 'none'
  element.style.overflow = 'visible'
  element.style.overflowY = 'visible'
  nextTick(async () => {
    const isDark = document.documentElement.getAttribute('theme-mode') === 'dark'
    try {
      const canvas = await html2canvas(element, {
        useCORS: true,
        scale: 2,
        allowTaint: true,
        logging: false,
        backgroundColor: isDark ? '#1e1e1e' : '#ffffff'
      })
      element.style.height = savedTargetStyle.height
      element.style.maxHeight = savedTargetStyle.maxHeight
      element.style.overflow = savedTargetStyle.overflow
      element.style.overflowY = savedTargetStyle.overflowY
      savedStyles.forEach(({ el, overflow, overflowY, height, maxHeight }) => {
        el.style.overflow = overflow
        el.style.overflowY = overflowY
        el.style.height = height
        el.style.maxHeight = maxHeight
      })
      const dataUrl = canvas.toDataURL('image/png')
      const base64 = dataUrl.replace(/^data:image\/png;base64,/, '')
      const result = await SaveImage(name + '[' + code + ']AI 分析', base64)
      if (result && !result.includes('异常') && !result.includes('无法')) {
        message.success('已导出为 PNG 图片：' + result)
      } else {
        message.info(result || '导出取消')
      }
    } catch (e) {
      element.style.height = savedTargetStyle.height
      element.style.maxHeight = savedTargetStyle.maxHeight
      element.style.overflow = savedTargetStyle.overflow
      element.style.overflowY = savedTargetStyle.overflowY
      savedStyles.forEach(({ el, overflow, overflowY, height, maxHeight }) => {
        el.style.overflow = overflow
        el.style.overflowY = overflowY
        el.style.height = height
        el.style.maxHeight = maxHeight
      })
      message.error('导出图片失败：' + (e?.message ?? e))
    }
  })
}

async function copyToClipboard() {
  try {
    await navigator.clipboard.writeText(data.airesult);
    message.success('分析结果已复制到剪切板');
  } catch (err) {
    message.error('复制失败：' + err);
  }
}

function saveAsMarkdown() {
  SaveAsMarkdown(data.code, data.name).then(result => {
    message.success(result)
  })
}

function getHtml(ref) {
  if (ref.value) {
    const rootElement = ref.value.$el;
    return rootElement.innerHTML;
  } else {
    console.error('ref is not yet available');
    return "";
  }
}

async function saveAsWord() {
  const html = getHtml(mdPreviewRef)
  const tipsHtml = getHtml(tipsRef)
  const value = `
         ${html}
         <hr>
         <div style="font-size: 12px;color: red">
         ${tipsHtml}
          </div>
<br>
本报告由 go-stock 项目生成：
<p>
<a href="https://github.com/ArvinLovegood/go-stock">
AI 赋能股票分析：自选股行情获取，成本盈亏展示，涨跌报警推送，市场整体/个股情绪分析，K 线技术指标分析等。数据全部保留在本地。支持 DeepSeek，OpenAI，Ollama，LMStudio，AnythingLLM，硅基流动，火山方舟，阿里云百炼等平台或模型。
</a></p>
`
  const blob = await asBlob(value, {orientation: 'portrait'})
  const {platform} = await Environment()
  switch (platform) {
    case 'windows':
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = `${data.name}[${data.code}]-ai-analysis-result.docx`;
      a.click()
      URL.revokeObjectURL(a.href);
      a.remove()
      break
    default:
      const arrayBuffer = await blob.arrayBuffer()
      const uint8Array = new Uint8Array(arrayBuffer)
      const binary = uint8Array.reduce((data, byte) => data + String.fromCharCode(byte), '')
      const base64 = btoa(binary)
      await SaveWordFile(`${data.name}[${data.code}]-ai-analysis-result.docx`, base64).then(result => {
        message.success(result)
      })
  }
}

function share(code, name) {
  ShareAnalysis(code, name).then(msg => {
    notify.info({
      avatar: () =>
          h(NAvatar, {
            size: 'small',
            round: false,
            src: icon.value
          }),
      title: '分享到社区',
      duration: 1000 * 30,
      content: () => {
        return h('div', {
          style: {
            'text-align': 'left',
            'font-size': '14px',
          }
        }, {default: () => msg})
      },
    })
  })
}
</script>

<template>
  <vue-danmaku v-model:danmus="danmus" useSlot style="height:100px; width:100%;z-index: 9;position:absolute; top: 400px; pointer-events: none;">
    <template v-slot:dm="{ danmu }">
      <n-gradient-text type="info">
        <n-icon :component="ChatboxOutline"/>{{ danmu }}
      </n-gradient-text>
    </template>
  </vue-danmaku>

  <n-divider style="margin: 4px 0 8px 0"/>

  <n-grid :x-gap="10" :y-gap="10" :cols="2" responsive="screen" item-responsive>
    <n-gi v-for="info in followList" :key="info.code" :id="info.code + '_gi'">
      <n-card :id="info.code" size="small" hoverable>
        <template #header>
          <n-text style="font-size: 15px; font-weight: 600;">{{ info.fundBasic?.fullName || info.name }}</n-text>
        </template>
        <template #header-extra>
          <n-flex :wrap="false" align="center" :size="4">
            <n-tag size="small" :bordered="false" type="info">{{ info.code }}</n-tag>
            <n-tag size="small" :bordered="false" type="warning">{{ info.fundBasic?.type || '' }}</n-tag>
          </n-flex>
        </template>

        <n-grid :cols="24" :x-gap="16">
          <n-gi :span="10">
            <n-flex align="center" :size="12" :wrap="false">
              <div v-if="!isOnExchangeFund(info.code) && info.netActualRate != null" style="min-width: 100px;">
                <div style="font-size: 12px; color: #999;">实际净值</div>
                <n-text :type="rateType(info.netActualRate)" style="font-size: 22px; font-weight: 700;">
                  {{ info.netUnitValue }}
                </n-text>
                <n-text :type="rateType(info.netActualRate)" style="font-size: 14px; margin-left: 4px;">
                  {{ info.netActualRate > 0 ? '+' : '' }}{{ info.netActualRate.toFixed(2) }}%
                </n-text>
                <div style="font-size: 11px; color: #999;">{{ info.netUnitValueDate }}</div>
              </div>
              <template v-else>
                <div v-if="info.netEstimatedUnit || info.fundBasic?.netEstimatedUnit" style="min-width: 100px;">
                  <div style="font-size: 12px; color: #999;">{{ isOnExchangeFund(info.code) ? '实时价格' : '估算净值' }}</div>
                  <n-text :type="rateType(info.netEstimatedRate || info.fundBasic?.netEstimatedRate)" style="font-size: 22px; font-weight: 700;">
                    {{ info.netEstimatedUnit || info.fundBasic?.netEstimatedUnit }}
                  </n-text>
                  <n-text :type="rateType(info.netEstimatedRate || info.fundBasic?.netEstimatedRate)" style="font-size: 14px; margin-left: 4px;">
                    {{ (info.netEstimatedRate || info.fundBasic?.netEstimatedRate) > 0 ? '+' : '' }}{{ (info.netEstimatedRate || info.fundBasic?.netEstimatedRate)?.toFixed(2) }}%
                  </n-text>
                </div>
                <div v-else-if="info.netUnitValue || info.fundBasic?.netUnitValue" style="min-width: 100px;">
                  <div style="font-size: 12px; color: #999;">单位净值</div>
                  <n-text style="font-size: 22px; font-weight: 700;">{{ info.netUnitValue || info.fundBasic?.netUnitValue }}</n-text>
                </div>
                <n-divider vertical v-if="(info.netEstimatedUnit || info.fundBasic?.netEstimatedUnit) && (info.netUnitValue || info.fundBasic?.netUnitValue)"/>
                <div v-if="(info.netUnitValue || info.fundBasic?.netUnitValue) && (info.netEstimatedUnit || info.fundBasic?.netEstimatedUnit)">
                  <div style="font-size: 12px; color: #999;">单位净值</div>
                  <n-text style="font-size: 15px;">{{ info.netUnitValue || info.fundBasic?.netUnitValue }}</n-text>
                  <div style="font-size: 11px; color: #999;">{{ info.netUnitValueDate || info.fundBasic?.netUnitValueDate }}</div>
                </div>
              </template>
            </n-flex>

            <n-flex :size="4" style="margin-top: 8px;" :wrap="true">
              <n-tag size="tiny" :type="growthType(info.fundBasic?.netGrowth1)" :bordered="false" v-if="info.fundBasic?.netGrowth1">近1月 {{ info.fundBasic.netGrowth1 }}%</n-tag>
              <n-tag size="tiny" :type="growthType(info.fundBasic?.netGrowth3)" :bordered="false" v-if="info.fundBasic?.netGrowth3">近3月 {{ info.fundBasic.netGrowth3 }}%</n-tag>
              <n-tag size="tiny" :type="growthType(info.fundBasic?.netGrowth6)" :bordered="false" v-if="info.fundBasic?.netGrowth6">近6月 {{ info.fundBasic.netGrowth6 }}%</n-tag>
              <n-tag size="tiny" :type="growthType(info.fundBasic?.netGrowth12)" :bordered="false" v-if="info.fundBasic?.netGrowth12">近1年 {{ info.fundBasic.netGrowth12 }}%</n-tag>
              <n-tag size="tiny" :type="growthType(info.fundBasic?.netGrowth36)" :bordered="false" v-if="info.fundBasic?.netGrowth36">近3年 {{ info.fundBasic.netGrowth36 }}%</n-tag>
              <n-tag size="tiny" :type="growthType(info.fundBasic?.netGrowth60)" :bordered="false" v-if="info.fundBasic?.netGrowth60">近5年 {{ info.fundBasic.netGrowth60 }}%</n-tag>
              <n-tag size="tiny" :type="growthType(info.fundBasic?.netGrowthYTD)" :bordered="false" v-if="info.fundBasic?.netGrowthYTD">今年来 {{ info.fundBasic.netGrowthYTD }}%</n-tag>
              <n-tag size="tiny" :type="growthType(info.fundBasic?.netGrowthAll)" :bordered="false" v-if="info.fundBasic?.netGrowthAll">成立来 {{ info.fundBasic.netGrowthAll }}%</n-tag>
            </n-flex>
          </n-gi>

          <n-gi :span="14">
            <div v-if="holdingsMap[info.code] && holdingsMap[info.code].length" class="holdings-panel">
              <div class="holdings-title">
                十大持仓
                <n-text v-if="holdingsMap[info.code][0]?.quarter" depth="3" style="font-size: 11px; margin-left: 4px;">
                  {{ holdingsMap[info.code][0].quarter }}
                </n-text>
              </div>
              <div class="holdings-cols">
                <template v-for="(half, hi) in splitHalves(holdingsMap[info.code])" :key="hi">
                  <div class="holdings-col">
                    <div class="holdings-header">
                      <span>名称</span>
                      <span>占比</span>
                      <span>最新价</span>
                      <span>涨跌幅</span>
                    </div>
                    <div v-for="stock in half" :key="stock.stockCode" class="holding-row">
                      <span class="holding-name" :title="stock.stockName">{{ stock.stockName }}</span>
                      <span class="holding-ratio" :style="{ color: ratioColor(stock.ratio) }">{{ stock.ratio?.toFixed(2) }}%</span>
                      <span class="holding-price" :style="{ color: changeColor(stock.changeRate) }">{{ stock.price ? stock.price.toFixed(2) : '-' }}</span>
                      <span class="holding-change" :style="{ color: changeColor(stock.changeRate) }">{{ changeText(stock.changeRate) }}</span>
                    </div>
                  </div>
                </template>
              </div>
            </div>
            <div v-else class="holdings-panel">
              <n-text depth="3" style="font-size: 12px;">暂无持仓数据</n-text>
            </div>
          </n-gi>
        </n-grid>

        <template #footer>
          <n-flex justify="space-between" align="center">
            <n-text depth="3" style="font-size: 12px;">
              {{ info.fundBasic?.company }} · {{ info.fundBasic?.manager }}
            </n-text>
          </n-flex>
        </template>

        <template #action>
          <n-flex justify="space-between" align="center">
            <n-text depth="3" style="font-size: 11px;">{{ countdown }}s 后刷新</n-text>
            <n-flex :size="8">
              <n-button size="tiny" :loading="refreshing" @click="manualRefresh">
                <template #icon><n-icon :component="RefreshOutline"/></template>
              </n-button>
              <n-button v-if="data.openAiEnable" size="tiny" type="warning" @click="aiCheckFund(info.name, info.code)">AI 分析</n-button>
              <n-button size="tiny" type="error" @click="showChart(info.code, info.name)">历史净值</n-button>
              <n-button size="tiny" type="warning" @click="search(info.code)">详情</n-button>
              <n-button size="tiny" @click="unFollow(info.code)">取消关注</n-button>
            </n-flex>
          </n-flex>
        </template>
      </n-card>
    </n-gi>
  </n-grid>



  <n-modal
    v-model:show="chartModalShow"
    :title="chartFundName + ' - ' + chartFundCode"
    preset="card"
    style="width: 90vw; max-width: 1100px;"
    :mask-closable="true"
  >
    <FundKlineChart
      v-if="chartFundCode"
      :key="chartFundCode"
      :fund-code="chartFundCode"
      :fund-name="chartFundName"
      :dark-theme="darkTheme"
      :chart-height="400"
    />

    <n-divider style="margin: 12px 0 8px 0">{{ isOnExchangeFund(chartFundCode) ? '历史行情' : '历史净值' }}</n-divider>

    <n-data-table
      :columns="netValueColumns"
      :data="netValueData"
      :loading="netValueLoading"
      :pagination="{ pageSize: 10 }"
      size="small"
      :bordered="false"
      :max-height="300"
      striped
    />
  </n-modal>

  <n-modal transform-origin="center" v-model:show="modalShow4" preset="card" style="width: max(30%, 400px);max-width: calc(100vw - 32px);"
           :title="'['+data.name+']AI 分析'">
    <n-spin size="small" :show="data.loading">
      <MdEditor v-if="enableEditor" :toolbars="toolbars" ref="mdEditorRef" style="height: 440px;max-height: 60vh;text-align: left"
                :modelValue="data.airesult" :theme="theme">
        <template #defToolbars>
          <ExportPDF :file-name="data.name+'['+data.code+']AI 分析报告'" style="text-align: left"
                     :modelValue="data.airesult" @onProgress="handleProgress"/>
        </template>
      </MdEditor>
      <div v-if="!enableEditor" ref="aiResultScrollRef" style="height: 440px;max-height: 60vh;text-align: left;overflow-y: auto;">
        <MdPreview ref="mdPreviewRef" :modelValue="data.airesult" :theme="theme"/>
      </div>
    </n-spin>
    <template #footer>
      <n-flex justify="space-between" ref="tipsRef">
        <n-text type="info" v-if="data.time">
          <n-tag v-if="data.modelName" type="warning" round :title="data.chatId" :bordered="false">
            {{ data.modelName }}
          </n-tag>
          {{ data.time }}
        </n-text>
        <n-text type="error">*AI 分析结果仅供参考，请以实际行情为准。投资需谨慎，风险自担。</n-text>
      </n-flex>
    </template>
    <template #action>
      <n-flex justify="left" style="margin-bottom: 10px">
        <n-switch v-model:value="enableTools" :round="false">
          <template #checked>
            工具调用
          </template>
          <template #unchecked>
            非工具调用
          </template>
        </n-switch>
        <n-switch v-model:value="thinkingMode" :round="false">
          <template #checked>
            思考模式
          </template>
          <template #unchecked>
            非思考模式
          </template>
        </n-switch>
        <n-gradient-text type="error" style="margin-left: 10px">
          *AI 函数工具调用可以增强 AI 获取数据的能力，但会消耗更多 tokens。
        </n-gradient-text>
      </n-flex>
      <n-flex justify="space-between" style="margin-bottom: 10px">
        <n-select style="width: 31%" v-model:value="data.aiConfigId" label-field="name" value-field="ID"
                  :options="aiConfigs" placeholder="请选择 AI 模型服务配置"/>
        <n-select style="width: 31%" v-model:value="data.sysPromptId" label-field="name" value-field="ID"
                  :options="sysPromptOptions" placeholder="请选择系统提示词"/>
        <n-select style="width: 31%" v-model:value="data.question" label-field="name" value-field="content"
                  :options="userPromptOptions" placeholder="请选择用户提示词"/>
      </n-flex>
      <n-flex justify="right">
        <n-input v-model:value="data.question" style="text-align: left" clearable
                 type="textarea"
                 :show-count="true"
                 placeholder="请输入您的问题：例如{{fundName}}[{{fundCode}}] 分析和总结"
                 :autosize="{
              minRows: 2,
              maxRows: 5
            }"
        />
        <n-button size="tiny" type="warning" @click="aiReCheckFund(data.name,data.code)">开始 AI 分析</n-button>
        <n-button size="tiny" type="info" @click="saveAsImage(data.name,data.code)">保存为图片</n-button>
        <n-button size="tiny" type="success" @click="copyToClipboard">复制到剪切板</n-button>
        <n-button size="tiny" type="primary" @click="saveAsMarkdown">保存为 Markdown 文件</n-button>
        <n-button size="tiny" type="primary" @click="saveAsWord">保存为 Word 文件</n-button>
        <n-button size="tiny" type="error" @click="share(data.code,data.name)">分享到项目社区</n-button>
      </n-flex>
    </template>
  </n-modal>

  <div style="position: fixed;bottom: 18px;right:5px;z-index: 10;width: 400px">
    <n-input-group>
      <n-auto-complete
        v-model:value="data.name"
        :input-props="{ autocomplete: 'disabled' }"
        :options="options"
        placeholder="基金名称/代码/弹幕"
        clearable
        @update-value="getFundList"
        :on-select="onSelectFund"
      />
      <n-popover trigger="manual" :show="showPopover">
        <template #trigger>
          <n-button type="primary" @click="AddFund">
            <n-icon :component="Add"/>&nbsp;关注
          </n-button>
        </template>
        <span>输入基金名称/代码关键词开始吧~~~</span>
      </n-popover>
      <n-button type="info" @click="SendDanmu" v-if="data.enableDanmu">
        <n-icon :component="ChatboxOutline"/>&nbsp;发送弹幕
      </n-button>
    </n-input-group>
  </div>
</template>

<style scoped>
.blink-border {
  animation: blink-border 1s linear infinite;
  border: 4px solid transparent;
}

@keyframes blink-border {
  0% { border-color: red; }
  50% { border-color: transparent; }
  100% { border-color: red; }
}

.holdings-panel {
  border-left: 1px solid var(--n-border-color, #efeff5);
  padding-left: 12px;
  height: 100%;
}

.holdings-title {
  font-size: 12px;
  font-weight: 600;
  color: #666;
  margin-bottom: 4px;
}

.holdings-cols {
  display: flex;
  gap: 10px;
}

.holdings-col {
  flex: 1;
  min-width: 0;
}

.holdings-header {
  display: grid;
  grid-template-columns: 1fr 44px 48px 52px;
  gap: 0 4px;
  font-size: 10px;
  color: var(--n-text-color-3, #999);
  padding-bottom: 2px;
  border-bottom: 1px solid var(--n-border-color, #efeff5);
  margin-bottom: 2px;
}

.holdings-header span:not(:first-child) {
  text-align: right;
}

.holding-row {
  display: grid;
  grid-template-columns: 1fr 44px 48px 52px;
  gap: 0 4px;
  align-items: center;
  font-size: 11px;
  line-height: 18px;
  white-space: nowrap;
}

.holding-name {
  overflow: hidden;
  text-overflow: ellipsis;
}

.holding-ratio {
  text-align: right;
  font-weight: 500;
}

.holding-price {
  text-align: right;
  color: var(--n-text-color-3, #999);
}

.holding-change {
  text-align: right;
  font-size: 11px;
}
</style>
