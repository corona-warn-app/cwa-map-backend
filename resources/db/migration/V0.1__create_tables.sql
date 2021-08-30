create table centers
(
    uuid           varchar(36)  not null primary key,
    operator_uuid  varchar(36)  not null,
    user_reference varchar(64),
    website        varchar(264),
    name           varchar(128) not null,
    address        varchar(264) not null,
    address_note   varchar(128),
    opening_hours  varchar(64)[],
    enter_date     timestamp,
    leave_date     timestamp,
    latitude       decimal      not null,
    longitude      decimal      not null,
    dcc            bool,
    appointment    varchar(32),
    test_kinds     varchar(64)[]

);

create index centers_appointment_index
    on centers (appointment);

create index centers_test_kinds_index
    on centers (test_kinds);

create index centers_dcc_index
    on centers (dcc);

create index centers_lat_lng_index
    on centers (latitude, longitude);

create table operators
(
    uuid            varchar(36)  not null primary key,
    subject         varchar(128),
    operator_number varchar(32),
    name            varchar(128) not null,
    logo            text,
    marker_icon     text
);

alter table centers
    add constraint centers_operators_uuid_fk
        foreign key (operator_uuid) references operators
            on update cascade on delete cascade;


