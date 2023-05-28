-- Migration history tracking table
CREATE TABLE schema_migrations_history
(
  id         SERIAL PRIMARY KEY NOT NULL,
  version    BIGINT             NOT NULL,
  applied_at timestamptz        NOT NULL DEFAULT NOW()
);

-- Migration history tracking function
CREATE OR REPLACE FUNCTION track_applied_migration()
  RETURNS TRIGGER AS
$$
DECLARE
  _current_version integer;
BEGIN
  SELECT COALESCE(MAX(version), 0) FROM schema_migrations_history INTO _current_version;
  IF new.version > _current_version THEN
    INSERT INTO schema_migrations_history(version) VALUES (new.version);
  END IF;
  RETURN NEW;
END;
$$ language 'plpgsql' STRICT;

-- Migration history tracking trigger
CREATE TRIGGER track_applied_migrations
  AFTER UPDATE
  ON schema_version
  FOR EACH ROW
EXECUTE PROCEDURE track_applied_migration();

-- Random int between low and high
CREATE OR REPLACE FUNCTION rand_between(low INTEGER, high INTEGER)
  RETURNS INTEGER AS
$$
BEGIN
  RETURN floor(random() * (high - low + 1) + low);
END;
$$ language 'plpgsql' STRICT;
