-- 0001_init.sql  (MySQL 8.0)
-- 对账系统初始化：任务、账单记录、差异、人工反馈
-- 约定：金额单位一律为「分」(BIGINT)；时间统一 DATETIME(3)

CREATE TABLE IF NOT EXISTS recon_task (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name        VARCHAR(64)     NOT NULL,
    status      VARCHAR(16)     NOT NULL DEFAULT 'CREATED', -- CREATED|PARSING|MATCHING|MATCHED|FAILED
    total_count INT             NOT NULL DEFAULT 0,
    diff_count  INT             NOT NULL DEFAULT 0,
    diff_amount BIGINT          NOT NULL DEFAULT 0,
    created_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    finish_time DATETIME(3)     NULL,
    PRIMARY KEY (id),
    KEY idx_task_status (status)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '对账任务';

CREATE TABLE IF NOT EXISTS recon_record (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id    BIGINT UNSIGNED NOT NULL,
    source     TINYINT         NOT NULL, -- 1=渠道 2=系统
    trade_no   VARCHAR(64)     NOT NULL DEFAULT '', -- 渠道交易号
    biz_no     VARCHAR(64)     NOT NULL DEFAULT '', -- 商户单号/系统支付单号
    amount     BIGINT          NOT NULL DEFAULT 0,  -- 分
    fee        BIGINT          NOT NULL DEFAULT 0,  -- 分
    trade_time DATETIME(3)     NULL,
    status     VARCHAR(32)     NOT NULL DEFAULT '',
    remark     VARCHAR(255)    NOT NULL DEFAULT '',
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_record_task_source (task_id, source),
    KEY idx_record_task_trade (task_id, trade_no),
    KEY idx_record_task_biz (task_id, biz_no),
    CONSTRAINT fk_record_task FOREIGN KEY (task_id) REFERENCES recon_task (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '账单原始记录（渠道与系统合并）';

CREATE TABLE IF NOT EXISTS recon_diff (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id           BIGINT UNSIGNED NOT NULL,
    channel_record_id BIGINT UNSIGNED NULL,
    system_record_id  BIGINT UNSIGNED NULL,
    diff_type         VARCHAR(32)     NOT NULL, -- channel_only|system_only|amount_mismatch
    amount_diff       BIGINT          NOT NULL DEFAULT 0,
    category          VARCHAR(32)     NOT NULL DEFAULT '', -- 六类之一，空=未归因
    confidence        DOUBLE          NOT NULL DEFAULT 0,
    reason            VARCHAR(512)    NOT NULL DEFAULT '',
    evidence          JSON            NULL,
    suggestion        VARCHAR(16)     NOT NULL DEFAULT '', -- confirm|investigate|ignore
    state             VARCHAR(16)     NOT NULL DEFAULT 'PENDING',
    fallback          TINYINT(1)      NOT NULL DEFAULT 0,
    ai_model          VARCHAR(64)     NOT NULL DEFAULT '',
    prompt_version    VARCHAR(32)     NOT NULL DEFAULT '',
    match_rule        VARCHAR(16)     NOT NULL DEFAULT '', -- exact|fuzzy|none
    created_at        DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at        DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_diff_task_state (task_id, state),
    KEY idx_diff_task_category (task_id, category),
    CONSTRAINT fk_diff_task FOREIGN KEY (task_id) REFERENCES recon_task (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '对账差异';

CREATE TABLE IF NOT EXISTS diff_feedback (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    diff_id        BIGINT UNSIGNED NOT NULL,
    ai_category    VARCHAR(32)     NOT NULL DEFAULT '',
    human_category VARCHAR(32)     NOT NULL,
    operator       VARCHAR(64)     NOT NULL DEFAULT '',
    created_at     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_feedback_diff (diff_id),
    CONSTRAINT fk_feedback_diff FOREIGN KEY (diff_id) REFERENCES recon_diff (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '人工纠正记录，用于 AI 评测与规则沉淀';
