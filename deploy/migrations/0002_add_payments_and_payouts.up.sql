-- =========================================
-- Дополняем rentals: предоплата и залог
-- =========================================

ALTER TABLE rentals
    ADD COLUMN prepayment          NUMERIC(12,2),  -- предоплата (nullable)
    ADD COLUMN deposit             NUMERIC(12,2),  -- залог (nullable)
    ADD COLUMN deposit_returned    NUMERIC(12,2),  -- сколько залога вернули (nullable)
    ADD COLUMN is_deposit_returned BOOLEAN NOT NULL DEFAULT FALSE;

-- =========================================
-- Выплаты партнёру
-- =========================================

CREATE TABLE payouts (
                         id              BIGSERIAL PRIMARY KEY,
                         period_start    DATE          NOT NULL,
                         period_end      DATE          NOT NULL,
                         total_revenue   NUMERIC(12,2) NOT NULL,
                         partner_share   NUMERIC(12,2) NOT NULL,
                         is_paid         BOOLEAN       NOT NULL DEFAULT FALSE,
                         paid_at         TIMESTAMPTZ,
                         comment         TEXT,
                         created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

                         CONSTRAINT chk_period CHECK (period_end >= period_start)
);

CREATE INDEX idx_payouts_period ON payouts(period_start, period_end);