create table slide
(
    id      serial primary key,
    bot_id  integer                                     not null
        constraint slide_bot_id_fk_channel_id
        references channel deferrable initially deferred,
    url     varchar(1024) default ''::character varying not null,
    page    integer                                     not null,
    current boolean                                     not null
);