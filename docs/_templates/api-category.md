<!--
模板：单分类 API 契约
用法：复制到 docs/api/tools/<分类>.md
-->

# {{分类中文名}} API

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：{{日期}}

覆盖 **{{N}}** 个{{分类中文名}}类工具。端点均为 `POST /api/v1/tools/{slug}`，请求与响应使用[统一信封](../README.md)。

| slug | 名称 | 传输 | 档位 |
| --- | --- | --- | --- |
| [{{slug}}](#{{slug}}) | {{名称}} | {{json\|multipart\|websocket}} | {{档位}} |

---

## {{slug}}

{{一句话说明}}

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `{{field}}` | {{string\|object\|int\|bool\|array}} | {{是\|否}} | {{说明}} |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `{{field}}` | {{类型}} | {{说明}} |

### 错误

`{{ERROR_CODE}}`、`{{ERROR_CODE}}`。

### 警告

`{{WARNING_CODE}}` —— {{说明}}。

### 示例

```json
{
  "{{field}}": "{{value}}"
}
```

```json
{
  "code": 0,
  "message": "ok",
  "data": { "{{field}}": "{{value}}" },
  "error": null,
  "meta": { "durationMs": 1, "inputBytes": 64, "outputBytes": 128, "warnings": [] },
  "requestId": "…"
}
```

> {{该工具的要点说明：默认值、边界、注意事项}}

---

## 相关文档

- [../README.md](../README.md) —— 通用信封、通用类型、限流
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/04-errors.md](../../standards/04-errors.md) —— 错误处理规范

<!--
检查清单：
- [ ] 工具数量与 catalog 一致
- [ ] 每个工具都有请求 / 响应 / 错误
- [ ] 字段名与实现一致
- [ ] 错误码在 docs/api/errors.md 中已收录
- [ ] 响应示例使用统一信封
-->
