-- Добавляем два новых поля в payouts
ALTER TABLE payouts
    ADD COLUMN owner_share        NUMERIC(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN amortization_share NUMERIC(12,2) NOT NULL DEFAULT 0;

-- Старые строки (если были) — заполним осмысленными значениями:
-- partner_share остаётся 10%, остальное — амортизация (по старым правилам 90% owner)
UPDATE payouts
SET owner_share        = total_revenue - partner_share,
    amortization_share = 0
WHERE owner_share = 0 AND amortization_share = 0;