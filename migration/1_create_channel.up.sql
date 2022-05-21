create table channel
(
    id                      serial
        constraint channel_pk
            primary key,
    name                    varchar(255)             default ''::character varying not null,
    external_channel_id     varchar(255)                                           not null,
    external_channel_secret text                                                   not null,
    access_token            text                                                   not null,
    access_token_expired_at timestamp with time zone                               not null,
    created_at              timestamp with time zone default now()                 not null,
    updated_at              timestamp with time zone default now()                 not null
);

create unique index channel_external_channel_id_uniq
    on channel (external_channel_id);
