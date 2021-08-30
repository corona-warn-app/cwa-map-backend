CREATE OR REPLACE FUNCTION haversine(lat1 double precision, lng1 double precision, lat2 double precision,
                                     lng2 double precision)
    RETURNS double precision AS
$BODY$
SELECT asin(
               sqrt(
                           sin(radians($3 - $1) / 2) ^ 2 +
                           sin(radians($4 - $2) / 2) ^ 2 *
                           cos(radians($1)) *
                           cos(radians($3))
                   )
           ) * 12742 AS distance;
$BODY$
    LANGUAGE sql IMMUTABLE
                 COST 100;