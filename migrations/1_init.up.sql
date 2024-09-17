create table if not exists users
(
    tg_id    integer primary key,
    limits   float not null default 0
);

comment on table users is 'Пользователи ТГ';

create unique index if not exists users_tg_id
    on users (tg_id);

create table if not exists usercategories
(
    id          serial primary key,
    short_name  text not null,
    name        text not null,
    description text not null,
    data_format text,
    price       float  not null
);

comment on table usercategories is 'Категории расходов';

create unique index if not exists usercategories_lower_name
    on usercategories (lower(name));

create table if not exists usermoneytransactions
(
    id          serial primary key,
    tg_id       integer     not null references users (tg_id) on delete cascade,
    category_id integer     not null references usercategories (id) on delete cascade,
    "timestamp" timestamp not null default current_timestamp
);

comment on table usermoneytransactions is 'Записи пользователей о расходах';

create index if not exists usermoneytransactions_tg_id_period
    on usermoneytransactions (tg_id, "timestamp");

create table if not exists userrefills
(
    id          serial    primary key,
    tg_id       integer   not null references users (tg_id) on delete cascade,
    invoice_id  integer   not null,
    status      text      not null default 'active',
    amount      float     not null,
    "timestamp" timestamp not null default current_timestamp
);

comment on table usermoneytransactions is 'Записи пользователей о пополнениях';

create index if not exists userrefills_tg_id_period
    on userrefills (tg_id, "timestamp");

create table if not exists workers
(
    tg_id    integer primary key,
    name     text not null,
    is_admin boolean not null default false,
    status   boolean not null default false
);

comment on column workers.tg_id is 'Работники ТГ';

create unique index if not exists workers_tg_id
    on workers (tg_id);

create table if not exists tickets
(
    id              serial primary key,
    worker_tg_id    integer not null references workers (tg_id) on delete cascade,
    buyer_tg_id     integer not null references users (tg_id) on delete cascade,
    category_id     integer not null references usercategories (id) on delete cascade,
    status          text not null default 'in_progress',
    "timestamp"     timestamp not null default current_timestamp
);

comment on table tickets is 'Тикеты';

create unique index if not exists tickets_id
    on tickets (id);
