<template>
  <div class="heatmap-container" :class="{ 'dark': darkTheme }">
    <div class="heatmap-toolbar">
      <div class="toolbar-title">
        <span>📊</span> 板块资金流向
      </div>
      <div class="view-toggle">
        <button 
          :class="{ active: viewMode === 'treemap' }" 
          @click="setViewMode('treemap')"
          title="以矩阵面积展示资金大小，无遮挡"
        >
          <span class="icon">🎚️</span> 矩阵热力图
        </button>
        <button 
          :class="{ active: viewMode === 'bubble' }" 
          @click="setViewMode('bubble')"
          title="以坐标轴展示市值与涨跌幅关系"
        >
          <span class="icon">💬</span> 气泡分布图
        </button>
      </div>
    </div>
    <div ref="heatmapRef" :style="{ width: '100%', height: (height - 54) + 'px' }"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import { GetSectorMoneyFlow, GetSectorTopStocks, IsTradingTime } from '../../wailsjs/go/main/App'
import { useMessage } from 'naive-ui'

const props = defineProps({
  darkTheme: {
    type: Boolean,
    default: false
  },
  height: {
    type: Number,
    default: 600
  }
})

const heatmapRef = ref(null)
const viewMode = ref('treemap') // 'treemap' or 'bubble'
let chartInstance = null
let updateTimer = null
const message = useMessage()

const setViewMode = (mode) => {
  viewMode.value = mode
}

const initChart = () => {
  if (heatmapRef.value) {
    if (chartInstance) {
      chartInstance.dispose()
    }
    chartInstance = echarts.init(heatmapRef.value, props.darkTheme ? 'dark' : 'light')
    
    const isTreemap = viewMode.value === 'treemap'
    
    // Configure initial options based on viewMode
    const option = {
      backgroundColor: 'transparent',
      grid: isTreemap ? null : {
        left: '4%',
        right: '8%',
        bottom: '10%',
        top: '12%',
        containLabel: true
      },
      tooltip: {
        trigger: 'item',
        enterable: true,
        backgroundColor: props.darkTheme ? 'rgba(30, 30, 35, 0.98)' : 'rgba(255, 255, 255, 0.98)',
        borderColor: props.darkTheme ? '#e06666' : '#1890ff',
        borderWidth: 1.5,
        borderRadius: 8,
        shadowBlur: 10,
        shadowColor: 'rgba(0, 0, 0, 0.2)',
        padding: 12,
        formatter: function (params) {
          return new Promise(async (resolve) => {
            const data = params.data;
            if (!data || !data.code) {
              resolve('');
              return;
            }
            let loadingHtml = `
              <div style="padding:4px; font-family: system-ui, -apple-system, sans-serif; line-height: 1.6;">
                <b style="font-size:14px; color:${props.darkTheme ? '#ffffff' : '#111111'}">${data.name}</b><br/>
                涨跌幅: <span style="font-weight:bold; color:${data.change > 0 ? '#ff4d4f' : '#52c41a'}">${data.change > 0 ? '+' : ''}${data.change}%</span><br/>
                主力净流: <span style="font-weight:bold; color:${data.inflow > 0 ? '#ff4d4f' : '#52c41a'}">${data.inflow > 0 ? '+' : ''}${(data.inflow / 100000000).toFixed(2)}亿</span><br/>
                <span style="color:#999; font-size:11px;">加载领涨股中...</span>
              </div>
            `;
            try {
              const res = await GetSectorTopStocks(data.code, 3);
              let topStocksHtml = '';
              if (res && res.data && res.data.diff) {
                res.data.diff.forEach(stock => {
                  const sChange = parseFloat(stock.f3) || 0;
                  topStocksHtml += `
                    <div style="display:flex; justify-content:space-between; width:170px; margin-bottom:4px; font-size:12px;">
                      <span style="color:${props.darkTheme ? '#ddd' : '#555'}">${stock.f14}</span>
                      <span style="font-weight:bold; color:${sChange > 0 ? '#ff4d4f' : '#52c41a'}">${sChange > 0 ? '+' : ''}${sChange}%</span>
                    </div>`;
                });
              }
              const finalHtml = `
                <div style="padding:4px; font-family: system-ui, -apple-system, sans-serif; line-height: 1.6;">
                  <b style="font-size:14px; color:${props.darkTheme ? '#ffffff' : '#111111'}">${data.name}</b><br/>
                  涨跌幅: <span style="font-weight:bold; color:${data.change > 0 ? '#ff4d4f' : '#52c41a'}">${data.change > 0 ? '+' : ''}${data.change}%</span><br/>
                  总市值: <span style="font-weight:bold; color:${props.darkTheme ? '#eee' : '#333'}">${(data.marketCap / 100000000).toFixed(2)}亿</span><br/>
                  主力净流: <span style="font-weight:bold; color:${data.inflow > 0 ? '#ff4d4f' : '#52c41a'}">${data.inflow > 0 ? '+' : ''}${(data.inflow / 100000000).toFixed(2)}亿</span><br/>
                  <hr style="margin:8px 0; border: 0; border-top: 1px dashed ${props.darkTheme ? '#444' : '#ddd'};" />
                  <b style="color:${props.darkTheme ? '#aaaaaa' : '#666666'}; font-size:12px; display:block; margin-bottom:4px;">板块领涨前三:</b>
                  ${topStocksHtml || '<span style="color:#999; font-size:12px;">暂无数据</span>'}
                </div>
              `;
              resolve(finalHtml);
            } catch(e) {
              resolve(loadingHtml + '<br/><span style="color:#ff4d4f; font-size:11px;">获取失败</span>');
            }
          });
        }
      },
      xAxis: isTreemap ? undefined : {
        type: 'value',
        name: '总市值 (亿)',
        nameLocation: 'end',
        nameGap: 15,
        nameTextStyle: {
          fontFamily: 'system-ui, -apple-system, sans-serif',
          fontWeight: 'bold',
          color: props.darkTheme ? '#bbbbbb' : '#555555',
          fontSize: 12
        },
        scale: true,
        splitLine: {
          show: true,
          lineStyle: {
            color: props.darkTheme ? '#2a2a2e' : '#f0f0f2',
            type: 'dashed'
          }
        },
        axisLabel: {
          formatter: (val) => (val / 100000000).toFixed(0),
          fontFamily: 'Outfit, Inter, system-ui, sans-serif',
          fontWeight: '500',
          color: props.darkTheme ? '#aaaaaa' : '#666666',
          fontSize: 11
        },
        axisLine: {
          show: true,
          lineStyle: {
            color: props.darkTheme ? '#3e3e42' : '#cccccc',
            width: 1.5
          }
        },
        axisTick: {
          show: true,
          lineStyle: {
            color: props.darkTheme ? '#3e3e42' : '#cccccc'
          }
        }
      },
      yAxis: isTreemap ? undefined : {
        type: 'value',
        name: '涨跌幅 (%)',
        nameTextStyle: {
          fontFamily: 'system-ui, -apple-system, sans-serif',
          fontWeight: 'bold',
          color: props.darkTheme ? '#bbbbbb' : '#555555',
          fontSize: 12
        },
        scale: true,
        splitLine: {
          show: true,
          lineStyle: {
            color: props.darkTheme ? '#2a2a2e' : '#f0f0f2',
            type: 'dashed'
          }
        },
        axisLabel: {
          formatter: '{value}%',
          fontFamily: 'Outfit, Inter, system-ui, sans-serif',
          fontWeight: '500',
          color: props.darkTheme ? '#aaaaaa' : '#666666',
          fontSize: 11
        },
        axisLine: {
          show: true,
          lineStyle: {
            color: props.darkTheme ? '#3e3e42' : '#cccccc',
            width: 1.5
          }
        },
        axisTick: {
          show: true,
          lineStyle: {
            color: props.darkTheme ? '#3e3e42' : '#cccccc'
          }
        }
      },
      series: isTreemap ? [{
        type: 'treemap',
        roam: false,
        nodeClick: false,
        breadcrumb: { show: false },
        label: {
          show: true,
          position: 'inside',
          fontFamily: 'Outfit, Inter, system-ui, sans-serif',
          formatter: function (params) {
            const data = params.data;
            const inflowYuan = data.inflow;
            const inflowText = inflowYuan >= 0 
              ? `+${(inflowYuan / 100000000).toFixed(1)}亿` 
              : `${(inflowYuan / 100000000).toFixed(1)}亿`;
            return `{name|${data.name}}\n{val|${inflowText}  ${data.change > 0 ? '+' : ''}${data.change}%}`;
          },
          rich: {
            name: {
              fontSize: 13,
              fontWeight: 'bold',
              color: '#ffffff',
              padding: [0, 0, 4, 0]
            },
            val: {
              fontSize: 11,
              fontWeight: 'bold',
              color: '#f0f0f0'
            }
          }
        },
        itemStyle: {
          borderColor: props.darkTheme ? '#1e1e24' : '#ffffff',
          borderWidth: 2,
          gapWidth: 2
        },
        data: []
      }] : [{
        type: 'scatter',
        itemStyle: {
          opacity: 0.92,
          shadowBlur: 14,
          shadowOffsetX: 0,
          shadowOffsetY: 4,
          shadowColor: props.darkTheme ? 'rgba(0, 0, 0, 0.75)' : 'rgba(0, 0, 0, 0.16)'
        },
        label: {
          show: true,
          formatter: function (params) {
            const data = params.data;
            const inflowYuan = data.inflow;
            const inflowText = inflowYuan >= 0 
              ? `+${(inflowYuan / 100000000).toFixed(1)}亿` 
              : `${(inflowYuan / 100000000).toFixed(1)}亿`;
            return `${data.name}\n${inflowText}`;
          },
          position: 'top',
          distance: 8,
          // Custom distinct text color (Cyan in dark mode, Deep Navy Blue in light mode) for high readability
          color: props.darkTheme ? '#00f2fe' : '#0033aa',
          textBorderColor: props.darkTheme ? '#18181c' : '#ffffff',
          textBorderWidth: 3.5,
          fontSize: 11,
          fontWeight: 'bold',
          fontFamily: 'Outfit, Inter, system-ui, sans-serif',
          lineHeight: 14
        },
        labelLayout: {
          hideOverlap: true
        },
        data: []
      }]
    }
    
    chartInstance.setOption(option, true)
  }
}

const fetchData = async () => {
  try {
    const res = await GetSectorMoneyFlow();
    console.log("GetSectorMoneyFlow returned:", res);
    
    if (res && res.data && res.data.diff) {
      const validItems = res.data.diff.filter(item => item.f62 !== "-" && item.f3 !== "-");
      
      const items = validItems.map(item => {
        const f20 = parseFloat(item.f20) || 0;
        const f3 = parseFloat(item.f3) || 0;
        const f62 = parseFloat(item.f62) || 0;
        
        return {
          name: item.f14,
          code: item.f12,
          value: viewMode.value === 'treemap' ? Math.abs(f62) : [f20, f3, Math.abs(f62)], // Size = absolute cash flow for Treemap
          marketCap: f20,
          change: f3,
          inflow: f62
        }
      });

      // Sort items by absolute inflow value
      items.sort((a, b) => Math.abs(b.inflow) - Math.abs(a.inflow));
      
      // Limit count: Treemap fits more blocks easily (40); Bubble chart is kept to 22 to avoid crowding
      const limitCount = viewMode.value === 'treemap' ? 40 : 22;
      const topItems = items.slice(0, limitCount);

      // 1. Calculate color map bounds
      const maxInflow = Math.max(...topItems.filter(i => i.inflow > 0).map(i => i.inflow), 1);
      const maxOutflow = Math.max(...topItems.filter(i => i.inflow < 0).map(i => Math.abs(i.inflow)), 1);

      topItems.forEach(item => {
        const inflow = item.inflow;
        
        if (viewMode.value === 'treemap') {
          // Treemap styling: Gradient colors based on flow direction and magnitude
          if (inflow >= 0) {
            const ratio = inflow / maxInflow;
            // Rich red gradients with clear contrast
            item.itemStyle = {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: ratio > 0.65 ? '#ff4d4f' : ratio > 0.35 ? '#ff7875' : '#ffa39e' },
                { offset: 1, color: ratio > 0.65 ? '#cf1322' : ratio > 0.35 ? '#ad2102' : '#f5222d' }
              ])
            };
          } else {
            const ratio = Math.abs(inflow) / maxOutflow;
            // Rich green gradients with clear contrast
            item.itemStyle = {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: ratio > 0.65 ? '#52c41a' : ratio > 0.35 ? '#73d13d' : '#95de64' },
                { offset: 1, color: ratio > 0.65 ? '#135200' : ratio > 0.35 ? '#237804' : '#389e0d' }
              ])
            };
          }
        } else {
          // Bubble styling: gorgeous 3D radial gradients (Red for Inflow, Green for Outflow)
          item.itemStyle = {
            color: inflow >= 0 ? new echarts.graphic.RadialGradient(0.4, 0.3, 1, [{
              offset: 0,
              color: 'rgba(255, 120, 120, 0.95)'
            }, {
              offset: 0.7,
              color: 'rgba(230, 40, 40, 0.95)'
            }, {
              offset: 1,
              color: 'rgba(165, 10, 10, 0.98)'
            }]) : new echarts.graphic.RadialGradient(0.4, 0.3, 1, [{
              offset: 0,
              color: 'rgba(120, 240, 120, 0.95)'
            }, {
              offset: 0.7,
              color: 'rgba(40, 190, 40, 0.95)'
            }, {
              offset: 1,
              color: 'rgba(15, 110, 15, 0.98)'
            }])
          };
          
          // Symbol sizes normalized between 25px and 70px to avoid crowding
          const maxBubbleInflow = Math.max(...topItems.map(i => Math.abs(i.inflow)));
          const minBubbleInflow = Math.min(...topItems.map(i => Math.abs(i.inflow)));
          let size = 25;
          if (maxBubbleInflow > minBubbleInflow) {
            size = 25 + ((Math.abs(inflow) - minBubbleInflow) / (maxBubbleInflow - minBubbleInflow)) * 45;
          }
          item.symbolSize = size;
        }
      });

      chartInstance.setOption({
        series: [{
          data: topItems
        }]
      });
    }
  } catch (error) {
    console.error('Failed to fetch sector money flow:', error);
  }
}

onMounted(() => {
  initChart();
  fetchData();
  
  // Set interval for every 5 minutes (300000 ms)
  updateTimer = setInterval(async () => {
    try {
      const isTrading = await IsTradingTime();
      if (isTrading) {
        fetchData();
      }
    } catch (e) {
      console.error(e)
    }
  }, 300000);
  
  const resizeHandler = () => {
    chartInstance && chartInstance.resize()
  };
  window.addEventListener('resize', resizeHandler);
  
  // Attach handler to instance so we can remove it
  heatmapRef.value._resizeHandler = resizeHandler;
})

watch(() => props.darkTheme, () => {
  initChart();
  fetchData();
});

watch(() => props.height, () => {
  chartInstance && chartInstance.resize();
})

watch(viewMode, () => {
  initChart();
  fetchData();
})

onUnmounted(() => {
  if (updateTimer) {
    clearInterval(updateTimer);
  }
  if (heatmapRef.value && heatmapRef.value._resizeHandler) {
    window.removeEventListener('resize', heatmapRef.value._resizeHandler);
  }
  if (chartInstance) {
    chartInstance.dispose();
  }
})
</script>

<style scoped>
.heatmap-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  font-family: 'Outfit', 'Inter', system-ui, -apple-system, sans-serif;
  background: transparent;
  box-sizing: border-box;
}

.heatmap-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.45);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.02);
  transition: all 0.3s ease;
}

.dark .heatmap-toolbar {
  background: rgba(30, 30, 35, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.toolbar-title {
  font-size: 15px;
  font-weight: 700;
  color: #1a1a1a;
  letter-spacing: 0.5px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dark .toolbar-title {
  color: #ffffff;
}

.view-toggle {
  display: flex;
  background: rgba(0, 0, 0, 0.06);
  padding: 3px;
  border-radius: 8px;
  border: 1px solid rgba(0, 0, 0, 0.02);
}

.dark .view-toggle {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.02);
}

.view-toggle button {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: none;
  background: transparent;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  color: #555555;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  outline: none;
}

.dark .view-toggle button {
  color: #bbbbbb;
}

.view-toggle button:hover {
  color: #111111;
}

.dark .view-toggle button:hover {
  color: #ffffff;
}

.view-toggle button.active {
  background: #ffffff;
  color: #0052d4;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
}

.dark .view-toggle button.active {
  background: rgba(255, 255, 255, 0.12);
  color: #00f2fe;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.view-toggle button .icon {
  font-size: 11px;
}
</style>
