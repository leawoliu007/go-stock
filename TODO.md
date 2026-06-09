# TODO

- [ ] 名站优选页面更新。去除 摸鱼 、 消息墙 
- [ ] 批量导入自选
- [ok ] 提示词广场 VIP专属福利弹窗无法关闭
- [ok ] AI助手弹窗宽度太宽，1/3宽度即可
- [ ] 自选股备注。是只需要在自选股列表里加一行备注显示和编辑，还是希望备注能同步到后端 JSON/SQLite 长期保存？
- [ ] 自定义指标
- [ ] 自选股指标值自动排序

## 指标
**K线分析界面的指标是内置在前端的**，不是后端下载的。核心位置在：

- `frontend/src/components/kline/`
  - `calc.ts` — 指标计算逻辑
  - `constants.ts` — 指标常量
  - `indicators/toggle.ts` — 指标显示开关
  - `indicators/tips.ts` — 指标说明文案（内置了约 28 个指标）
- 主页面：`frontend/src/components/kline-analysis.vue`

内置指标包括：MA、EMA、BOLL、MACD、KDJ、RSI、ATR、VWAP、MFI、KAMA、Keltner、Supertrend、Ichimoku、CCI、TTM Squeeze、SAR、Donchian、ADX、Williams %R、StochRSI、CMF、Aroon、CMO、Force Index、Pivot Points、DEMA 等。

**可以新增**，但需要在前端代码层面扩展：
1. 在 `calc.ts` 中增加计算函数
2. 在 `constants.ts` 和 `toggle.ts` 中注册新指标
3. 在 `tips.ts` 中添加使用说明
4. 涉及具体图表绘制时，基于 `lightweight-charts` 库扩展

由于全部在前端，无需后端配合。


