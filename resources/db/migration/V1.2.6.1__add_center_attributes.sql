alter table centers
    add last_update timestamptz,
    add visible     bool default true;

update centers set visible = true;