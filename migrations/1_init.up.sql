create table if not exists users
(
    tg_id    SERIAL primary key,
    access   boolean not null default false,
    accessed_at date
);

comment on table users is 'Пользователи ТГ';

create unique index if not exists users_tg_id
    on users (tg_id);
