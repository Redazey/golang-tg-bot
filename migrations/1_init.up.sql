create table if not exists users
(
    id       serial primary key,
    tg_id    integer not null,
    limits   float not null
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
    user_id     integer     not null references users (id) on delete cascade,
    category_id integer     not null references usercategories (id) on delete cascade,
    sum         numeric     not null, -- (сумма в копейках)
    "timestamp" timestamp not null default current_timestamp
);

comment on table usermoneytransactions is 'Записи пользователей о расходах';

-- Индекс по пользователю и времени операции для ускорения поиска записей за определенный период.
create index if not exists usermoneytransactions_user_id_period
    on usermoneytransactions (user_id, "timestamp");
