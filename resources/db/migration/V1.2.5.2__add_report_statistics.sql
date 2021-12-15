create table report_statistics
(
    subject varchar(128) not null,
    count   integer      not null,
    constraint report_statistics_pk primary key (subject)
)