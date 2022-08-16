create table report_center_statistics
(
    operator_uuid varchar(36)  not null
        references operators on delete cascade on update cascade,
    center_uuid   varchar(36)  not null
        references centers on delete cascade on update cascade,
    subject       varchar(128) not null,
    count         integer      not null,
    constraint report_center_statistics_pk
        primary key (operator_uuid, center_uuid, subject)
);