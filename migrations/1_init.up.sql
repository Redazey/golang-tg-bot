create table if not exists users
(
    tg_id    integer primary key,
    limits   float not null default 0
);

comment on table users is 'Пользователи ТГ';

-- Индекс по ТГ-идентификатору пользователя для ускорения поиска.
create unique index if not exists users_tg_id
    on users (tg_id);

create table if not exists usercategories
(
    id      serial primary key,
    name    text    not null
        constraint usercategories_name_check
            check (name <> '')
);

comment on table usercategories is 'Категории расходов';

-- Индекс по пользователю и наименованию категории (lower для регистронезависимого поиска).
create unique index if not exists usercategories_lower_name
    on usercategories (lower(name));

create table if not exists usermoneytransactions
(
    id          serial primary key,
    tg_id     integer     not null references users (tg_id) on delete cascade,
    category_id integer     not null references usercategories (id) on delete cascade,
    sum         numeric     not null, -- (сумма в копейках)
    "timestamp" timestamp not null default current_timestamp
);

comment on table usermoneytransactions is 'Записи пользователей о расходах';

-- Индекс по пользователю и времени операции для ускорения поиска записей за определенный период.
create index if not exists usermoneytransactions_tg_id_period
    on usermoneytransactions (tg_id, "timestamp");

create table if not exists workers
(
    tg_id    integer primary key,
    is_admin boolean not null default false,
    status   boolean not null default false
);

comment on column workers.tg_id is 'Работники ТГ';
-- Индекс по ТГ-идентификатору пользователя для ускорения поиска.
create unique index if not exists workers_tg_id
    on workers (tg_id);

create table if not exists tickets
(
    id       serial primary key,
    worker_tg_id    integer not null references workers (tg_id) on delete cascade,
    buyer_tg_id integer not null references users (tg_id) on delete cascade,
    status   text not null default 'in_progress',
    "timestamp" timestamp not null default current_timestamp
);

comment on table tickets is 'Тикеты';
-- Индекс по ТГ-идентификатору пользователя для ускорения поиска.
create unique index if not exists tickets_id
    on tickets (id);
