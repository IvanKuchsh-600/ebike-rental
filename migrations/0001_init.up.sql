-- ВЕЛОСИПЕДЫ
CREATE TABLE bikes (
                       id              BIGSERIAL PRIMARY KEY,
                       serial_number   TEXT        NOT NULL UNIQUE,   -- инвентарный номер
                       model           TEXT,
                       is_rented       BOOLEAN     NOT NULL DEFAULT FALSE,
                       is_broken       BOOLEAN     NOT NULL DEFAULT FALSE,
                       comment         TEXT,
                       created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- АРЕНДЫ
CREATE TABLE rentals (
                         id              BIGSERIAL PRIMARY KEY,
                         bike_id         BIGINT        NOT NULL REFERENCES bikes(id),

                         renter_name     TEXT          NOT NULL,          -- имя арендатора
                         renter_phone    TEXT          NOT NULL,          -- телефон арендатора

                         started_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
                         planned_end_at  TIMESTAMPTZ   NOT NULL,          -- на какой срок сдан
                         returned_at     TIMESTAMPTZ,                     -- когда реально вернули
                         total_amount    NUMERIC(12,2) NOT NULL,          -- сколько получу за срок
                         is_paid         BOOLEAN       NOT NULL DEFAULT FALSE,  -- оплатил?
                         comment         TEXT,
                         created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- ПОЛОМКИ
CREATE TABLE breakdowns (
                            id            BIGSERIAL PRIMARY KEY,
                            bike_id       BIGINT        NOT NULL REFERENCES bikes(id),
                            reason        TEXT          NOT NULL,             -- причина поломки
                            cost          NUMERIC(12,2) NOT NULL DEFAULT 0,   -- стоимость поломки
                            broken_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(), -- срок поломки
                            is_fixed      BOOLEAN       NOT NULL DEFAULT FALSE, -- починен ли
                            fixed_at      TIMESTAMPTZ,
                            comment       TEXT,
                            created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- чтобы быстро искать аренды по телефону
CREATE INDEX idx_rentals_renter_phone ON rentals(renter_phone);