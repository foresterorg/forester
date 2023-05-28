CREATE TABLE IF NOT EXISTS random_names
(
  id    INTEGER NOT NULL UNIQUE,
  value text NOT NULL
);

-- Random string as "NAME SURNAME" from frequently occurring given names and surnames
-- from the 1990 US Census (public domain data):
--
-- * 256 (8 bits) unique male given names
-- * 256 (8 bits) unique female given names
-- * 65,536 (16 bits) unique surnames
-- * with over 120 gender-neutral given names
--
-- Given names were filtered to be 3-5 characters long, surnames 5-8 characters, therefore
-- generated names are never longer than 14 characters (5+1+8). This gives 33,554,432 (25 bits)
-- total of male and female name combinations.
CREATE OR REPLACE FUNCTION random_name()
  RETURNS TEXT AS
$$
DECLARE
  n1 TEXT;
  n2 TEXT;
BEGIN
  SELECT value INTO n1 FROM random_names WHERE id = (SELECT rand_between(1, 512));
  SELECT value INTO n2 FROM random_names WHERE id = (SELECT rand_between(513, 66048));
  RETURN n1 || ' ' || n2;
END;
$$ language 'plpgsql' STRICT;
