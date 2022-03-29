delete from report_statistics;

alter table report_statistics
    add operator_uuid varchar(36) not null;

alter table report_statistics
    add constraint report_statistics_operators_uuid_fk
        foreign key (operator_uuid) references operators
            on update cascade on delete cascade;

alter table report_statistics drop constraint report_statistics_pk;

alter table report_statistics
    add constraint report_statistics_pk
        primary key (operator_uuid, subject);
