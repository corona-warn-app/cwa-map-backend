alter table centers
    add message varchar(128);

alter table centers
    alter column address_note type text using address_note::text;