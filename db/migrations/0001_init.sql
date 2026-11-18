-- 0001_init.sql
-- 对账系统初始化：任务、账单记录、差异、人工反馈
-- 约定：金额单位一律为「分」(BIGINT)；时间统一 TIMESTAMPTZ

CREATE TABLE IF NOT EXISTS recon_task (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL,
    status      VARCHAR(16)  NOT NULL DEFAULT 'CREATED', -- CREATED|PARSING|MATCHING|MATCHED|FAILED
    total_count INT          NOT NULL DEFAULT 0,
    diff_count  INT          NOT NULL DEFAULT 0,
    diff_amount BIGINT       NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    finish_time TIMESTAMPTZ
);

COMMENT ON TABLE recon_task IS '对账任务';

CREATE TABLE IF NOT EXISTS recon_record (
    id         BIGSERIAL PRIMARY KEY,
    task_id    BIGINT       NOT NULL REFERENCES recon_task (id) ON DELETE CASCADE,
    source     SMALLINT     NOT NULL,                    -- 1=渠道 2=系统
    trade_no   VARCHAR(64)  NOT NULL DEFAULT '',         -- 渠道交易号
    biz_no     VARCHAR(64)  NOT NULL DEFAULT '',         -- 商户单号/系统支付单号
    amount     BIGINT       NOT NULL DEFAULT 0,          -- 分
    fee        BIGINT       NOT NULL DEFAULT 0,          -- 分
    trade_time TIMESTAMPTZ,
    status     VARCHAR(32)  NOT NULL DEFAULT '',
    remark     VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON TABLE recon_record IS '账单原始记录（渠道与系统合并，source 区分）';

CREATE INDEX IF NOT EXISTS idx_record_task_source ON recon_record (task_id, source);
CREATE INDEX IF NOT EXISTS idx_record_task_trade ON recon_record (task_id, trade_no);
CREATE INDEX IF NOT EXISTS idx_record_task_biz ON recon_record (task_id, biz_no);

CREATE TABLE IF NOT EXISTS recon_diff (
    id                BIGSERIAL PRIMARY KEY,
    task_id           BIGINT       NOT NULL REFERENCES recon_task (id) ON DELETE CASCADE,
    channel_record_id BIGINT,                            -- 可为空（系统单边）
    system_record_id  BIGINT,                            -- 可为空（渠道单边）
    diff_type         VARCHAR(32)  NOT NULL,             -- channel_only|system_only|amount_mismatch
    amount_diff       BIGINT       NOT NULL DEFAULT 0,   -- 分，代码计算
    category          VARCHAR(32)  NOT NULL DEFAULT '',  -- 六类之一，空=未归因
    confidence        DOUBLE PRECISION NOT NULL DEFAULT 0,
    reason            VARCHAR(512) NOT NULL DEFAULT '',
    evidence          JSONB        NOT NULL DEFAULT '[]'::jsonb,
    suggestion        VARCHAR(16)  NOT NULL DEFAULT '',  -- confirm|investigate|ignore
    state             VARCHAR(16)  NOT NULL DEFAULT 'PENDING', -- PENDING|CLASSIFIED|CONFIRMED|CORRECTED
    fallback          BOOLEAN      NOT NULL DEFAULT false,
    ai_model          VARCHAR(64)  NOT NULL DEFAULT '',
    prompt_version    VARCHAR(32)  NOT NULL DEFAULT '',
    match_rule        VARCHAR(16)  NOT NULL DEFAULT '',  -- exact|fuzzy|none
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON TABLE recon_diff IS '对账差异';

CREATE INDEX IF NOT EXISTS idx_diff_task_state ON recon_diff (task_id, state);
CREATE INDEX IF NOT EXISTS idx_diff_task_category ON recon_diff (task_id, category);

CREATE TABLE IF NOT EXISTS diff_feedback (
    id             BIGSERIAL PRIMARY KEY,
    diff_id        BIGINT      NOT NULL REFERENCES recon_diff (id) ON DELETE CASCADE,
    ai_category    VARCHAR(32) NOT NULL DEFAULT '',
    human_category VARCHAR(32) NOT NULL,
    operator       VARCHAR(64) NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE diff_feedback IS '人工纠正记录，用于 AI 评测与规则沉淀';

CREATE INDEX IF NOT EXISTS idx_feedback_diff ON diff_feedback (diff_id);
