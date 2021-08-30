alter table centers
add ranking double precision;

update centers set ranking = random();