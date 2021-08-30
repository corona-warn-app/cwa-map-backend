create table statistics
(
    time            timestamp primary key,
    operators_count integer not null,
    total_count     integer not null,
    dcc_count       integer not null
);

-- create procedure collect_usage_statistics()
--     language sql
-- as
-- $$
-- insert into statistics (time, operators_count, total_count, dcc_count)
-- SELECT now(),
--        (select count(*) from operators),
--        (select count(*) from centers),
--        (select count(*) from centers where dcc = true);
-- $$;
