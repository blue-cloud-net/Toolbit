-- 0001_init —— 初始表结构
--
-- 本版**不含**访客标识列：访客维度在 v0.9.0 的 0002 迁移中补齐（见 docs/roadmap.md §12）。

-- 迁移版本表。
CREATE TABLE IF NOT EXISTS schema_version (
    version    INTEGER NOT NULL,
    applied_at TEXT    NOT NULL
);

-- 工具调用明细。
CREATE TABLE IF NOT EXISTS tool_calls (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    ts          INTEGER NOT NULL,             -- 毫秒时间戳
    day         TEXT    NOT NULL,             -- YYYY-MM-DD
    slug        TEXT    NOT NULL,             -- 工具标识
    category    TEXT    NOT NULL,             -- 分类
    status      INTEGER NOT NULL,             -- HTTP 状态码
    ok          INTEGER NOT NULL,             -- 1 成功 / 0 失败
    error_code  TEXT    NOT NULL DEFAULT '',  -- 机器可读错误码，成功时为空
    duration_ms INTEGER NOT NULL DEFAULT 0,   -- 处理耗时
    req_bytes   INTEGER NOT NULL DEFAULT 0,   -- 请求体字节数
    resp_bytes  INTEGER NOT NULL DEFAULT 0    -- 响应体字节数
);

CREATE INDEX IF NOT EXISTS idx_tool_calls_day ON tool_calls (day);
CREATE INDEX IF NOT EXISTS idx_tool_calls_slug_day ON tool_calls (slug, day);
