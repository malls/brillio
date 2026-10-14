-- Listings schema (docs/SPEC.md section 4.1). Identity is (source, id).
-- Applied inside a transaction by Store.Migrate: no BEGIN or COMMIT here.

CREATE TABLE schema_migrations (
  filename   TEXT PRIMARY KEY,
  applied_at TEXT NOT NULL
);

CREATE TABLE listings (
  id          TEXT    NOT NULL,
  source      TEXT    NOT NULL,
  address     TEXT    NOT NULL,
  city        TEXT    NOT NULL,
  state       TEXT    NOT NULL,
  zip         TEXT    NOT NULL,
  price       REAL    NOT NULL,
  bedrooms    INTEGER NOT NULL,
  bathrooms   REAL    NOT NULL,
  sqft        INTEGER NOT NULL,
  latitude    REAL    NOT NULL,
  longitude   REAL    NOT NULL,
  listed_date TEXT    NOT NULL, -- YYYY-MM-DD
  status      TEXT    NOT NULL CHECK (status IN ('active', 'pending', 'sold')),
  description TEXT    NOT NULL,
  PRIMARY KEY (source, id)
);

CREATE INDEX listings_city_nocase ON listings (city COLLATE NOCASE);
CREATE INDEX listings_price ON listings (price);
CREATE INDEX listings_bedrooms ON listings (bedrooms);
CREATE INDEX listings_status ON listings (status);
