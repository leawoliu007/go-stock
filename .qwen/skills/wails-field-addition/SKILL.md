---
name: Wails 全栈字段添加模式
description: 在 Wails (Go + Vue) 应用中添加新字段/功能的完整流程和注意事项
source: auto-skill
extracted_at: '2026-06-09T13:47:19.058Z'
updated_at: '2026-06-09T14:08:40.205Z'
---

# Wails 全栈字段添加模式

## 概述

在 Wails 桌面应用（Go 后端 + Vue/React 前端）中添加新字段或功能时，需要按照特定顺序修改多个层级的文件。

## 完整流程

### 1. 后端数据模型层
**文件**: `backend/data/*_api.go`

- 在结构体中添加新字段（如 `Remark string`）
- GORM 会自动处理数据库迁移（如果使用了 AutoMigrate）
- 添加 setter/getter 方法，遵循现有命名规范

```go
// 示例：添加备注字段
type FollowedStock struct {
    // ... existing fields
    Remark string
}

func (receiver StockDataApi) SetStockRemark(remark, stockCode string) string {
    err := db.Dao.Model(&FollowedStock{}).Where("stock_code = ?", stockCode).Update("remark", remark).Error
    if err != nil {
        return "设置失败"
    }
    return "设置成功"
}
```

### 2. App 包装层
**文件**: `app.go` 或 `app_common.go`

- 在 App 结构体上添加方法，调用后端 API
- 这是 Wails 自动绑定的入口点

```go
func (a *App) SetStockRemark(remark, stockCode string) string {
    return data.NewStockDataApi().SetStockRemark(remark, stockCode)
}
```

### 3. 前端绑定层（自动生成或手动）
**文件**: `frontend/wailsjs/go/main/App.js` 和 `App.d.ts`

**通常由 Wails 自动生成**，但如果需要手动添加：

```javascript
// App.js
export function SetStockRemark(arg1, arg2) {
  return window['go']['main']['App']['SetStockRemark'](arg1, arg2);
}
```

```typescript
// App.d.ts
export function SetStockRemark(arg1:string,arg2:string):Promise<string>;
```

### 4. TypeScript 模型层
**文件**: `frontend/wailsjs/go/models.ts`

- 在对应的类中添加字段声明
- 在构造函数中添加字段赋值

```typescript
export class FollowedStock {
    // ... existing fields
    Remark: string;

    constructor(source: any = {}) {
        // ... existing assignments
        this.Remark = source["Remark"];
    }
}
```

### 5. 前端组件层
**文件**: `frontend/src/components/*.vue`

- 在 `formModel` 或 `ref` 中添加字段
- 在模板中添加显示/编辑控件
- 在保存方法中调用后端 API

```vue
<script setup>
const formModel = ref({
  // ... existing fields
  remark: "",
})

// 保存时调用
SetStockRemark(formModel.remark, code).then(res => {
  message.success(res)
})
</script>

<template>
  <!-- 显示 -->
  <n-text @dblclick="editRemark(code, remark)">
    {{ remark || '备注' }}
  </n-text>

  <!-- 编辑 -->
  <n-input v-model:value="formModel.remark" />
</template>
```

## 关键注意事项

### ⚠️ 文件编码问题

**绝对不要使用 PowerShell 修改包含非 ASCII 字符的文件**（如中文）。

PowerShell 的 `-replace` 操作会破坏 UTF-8 编码，导致：
- 中文字符变成乱码
- 构建失败（"Unterminated string literal"）
- 需要 `git checkout` 恢复文件

**正确做法**：使用 `edit` 工具或文本编辑器直接修改。

### ⚠️ 缩进匹配

`edit` 工具要求 `old_string` 的缩进必须**完全匹配**文件中的实际内容：
- 使用 `read_file` 查看实际缩进
- Tab 和空格不能混用
- 如果编辑失败，先用 `read_file` 确认精确的空白字符

### ⚠️ 数据库迁移

- GORM 的 `AutoMigrate` 会自动添加新列
- 但如果表已存在且数据量大，迁移可能很慢
- 生产环境建议使用手动迁移脚本

### ⚠️ 前端数据流

- `followList` 获取自选股列表（含备注等配置信息）
- `Greet` 获取实时行情数据
- `addStockFollowData` 函数将配置信息（如备注）合并到实时行情中
- `results` 计算属性基于合并后的数据构建卡片显示
- **重要**：如果新字段不在 `addStockFollowData` 中合并，则不会显示在卡片上

### ⚠️ 前端刷新

修改后端数据后，前端需要刷新列表：

```javascript
GetFollowList(groupId).then(result => {
  followList.value = result
})
```

## 验证步骤

1. **后端构建**: `go build .`
2. **前端构建**: `cd frontend && npm run build`
3. **运行测试**: 启动应用，验证：
   - 新字段能正确显示
   - 能成功保存到数据库
   - 重启后数据持久化

## 常见错误排查

| 错误 | 原因 | 解决方案 |
|------|------|---------|
| `Unterminated string literal` | models.ts 编码被破坏 | `git checkout` 恢复，用 edit 工具重新修改 |
| `SetStockRemark is not defined` | 忘记在 App.js 中添加绑定 | 检查 wailsjs/go/main/App.js |
| 字段值为 undefined | models.ts 构造函数未赋值 | 添加 `this.Remark = source["Remark"]` |
| 数据库列不存在 | GORM 未自动迁移 | 重启应用或手动执行 AutoMigrate |
| 前端显示旧数据 | 未在 `addStockFollowData` 中合并字段 | 更新合并函数 |
