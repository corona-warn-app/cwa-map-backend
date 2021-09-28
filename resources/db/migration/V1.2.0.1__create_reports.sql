create table bug_reports
(
    uuid           varchar(36) not null primary key,
    created        timestamptz not null,
    operator_uuid  varchar(36) not null,
    email          varchar     not null,
    center_uuid    varchar     not null,
    center_name    varchar     not null,
    center_address varchar     not null,
    subject        text        not null,
    message        text
);

alter table operators
    add email                varchar,
    add bug_reports_receiver varchar;

alter table centers
    add email varchar;