package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"brillio/backend/internal/domain"
)

// SQL text. Every statement and fragment is a constant; caller values only
// ever travel as "?" arguments.
const (
	upsertListing = `INSERT INTO listings (id, source, address, city, state, zip, price, bedrooms,
  bathrooms, sqft, latitude, longitude, listed_date, status, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (source, id) DO UPDATE SET
  address = excluded.address, city = excluded.city, state = excluded.state, zip = excluded.zip,
  price = excluded.price, bedrooms = excluded.bedrooms, bathrooms = excluded.bathrooms,
  sqft = excluded.sqft, latitude = excluded.latitude, longitude = excluded.longitude,
  listed_date = excluded.listed_date, status = excluded.status, description = excluded.description`

	// selectListings reads the columns in scanListing's order.
	selectListings = "SELECT id, source, address, city, state, zip, price, bedrooms, bathrooms, sqft, " +
		"latitude, longitude, listed_date, status, description FROM listings"

	// Search filter clauses (docs/SPEC.md section 6.1), in their fixed order.
	whereMinPrice    = "price >= ?"
	whereMaxPrice    = "price <= ?"
	whereMinBedrooms = "bedrooms >= ?"
	whereCity        = "city = ? COLLATE NOCASE"
	// instr, not LIKE, so "%" and "_" in a token match literally.
	whereKeywordToken = "instr(lower(description), ?) > 0"
	whereStatusIn     = "status IN ("
	whereStatusInEnd  = ")"

	sqlWhere        = " WHERE "
	sqlAnd          = " AND "
	placeholder     = "?"
	placeholderSep  = ", "
	orderBySourceID = " ORDER BY source, id"

	// selectCities picks, for each city compared case-insensitively, the row
	// inserted first (lowest rowid; an upsert's DO UPDATE keeps the rowid) as
	// the display casing, and sorts A to Z. The first-seen order depends on
	// the implicit rowid, which a VACUUM could renumber; nothing runs one.
	selectCities = `SELECT city FROM listings
WHERE rowid IN (SELECT min(rowid) FROM listings WHERE city <> '' GROUP BY city COLLATE NOCASE)
ORDER BY city COLLATE NOCASE`
)

// Upsert inserts or updates listings by (source, id) in one transaction and
// returns how many it wrote, which is len(listings). Any failure rolls back
// every row. It never deletes rows missing from listings.
func (s *Store) Upsert(ctx context.Context, listings []domain.Listing) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("upsert: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // ErrTxDone after Commit
	stmt, err := tx.PrepareContext(ctx, upsertListing)
	if err != nil {
		return 0, fmt.Errorf("upsert: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	for _, l := range listings {
		if _, err := stmt.ExecContext(ctx,
			l.ID, l.Source, l.Address, l.City, l.State, l.Zip,
			l.Price, l.Bedrooms, l.Bathrooms, l.Sqft, l.Latitude, l.Longitude,
			l.ListedDate.Format(domain.DateLayout), string(l.Status), l.Description,
		); err != nil {
			return 0, fmt.Errorf("upsert %s/%s: %w", l.Source, l.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("upsert: commit: %w", err)
	}
	return len(listings), nil
}

// Search returns the listings matching f, ordered by (source, id). It never
// returns a nil slice: no match is an empty one.
//
// Non-nil price and bedroom bounds apply inclusively (a pointer to 0 is a
// real bound). City is trimmed and matched exactly, ignoring ASCII case; ""
// means any city. Every non-blank keyword token must appear in the
// description, compared lowercased and literally. Statuses lists the
// statuses to include; empty means all. f.Keyword is not used: only
// KeywordTokens filter.
func (s *Store) Search(ctx context.Context, f domain.SearchFilter) ([]domain.Listing, error) {
	query, args := buildSearch(f)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer func() { _ = rows.Close() }()
	listings := []domain.Listing{}
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, fmt.Errorf("search: %w", err)
		}
		listings = append(listings, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	return listings, nil
}

// buildSearch returns the search query for f and its arguments. The query
// text is built only from the constant fragments above and "?" placeholders;
// every filter value is an argument.
func buildSearch(f domain.SearchFilter) (string, []any) {
	var (
		clauses []string
		args    []any
	)
	if f.MinPrice != nil {
		clauses = append(clauses, whereMinPrice)
		args = append(args, *f.MinPrice)
	}
	if f.MaxPrice != nil {
		clauses = append(clauses, whereMaxPrice)
		args = append(args, *f.MaxPrice)
	}
	if f.MinBedrooms != nil {
		clauses = append(clauses, whereMinBedrooms)
		args = append(args, *f.MinBedrooms)
	}
	if city := strings.TrimSpace(f.City); city != "" {
		clauses = append(clauses, whereCity)
		args = append(args, city)
	}
	for _, token := range f.KeywordTokens {
		token = strings.ToLower(strings.TrimSpace(token))
		if token == "" {
			continue
		}
		clauses = append(clauses, whereKeywordToken)
		args = append(args, token)
	}
	if len(f.Statuses) > 0 {
		marks := make([]string, len(f.Statuses))
		for i, status := range f.Statuses {
			marks[i] = placeholder
			args = append(args, string(status))
		}
		clauses = append(clauses, whereStatusIn+strings.Join(marks, placeholderSep)+whereStatusInEnd)
	}

	query := selectListings
	if len(clauses) > 0 {
		query += sqlWhere + strings.Join(clauses, sqlAnd)
	}
	return query + orderBySourceID, args
}

// scanListing scans the current row of a selectListings query. listed_date
// is stored as YYYY-MM-DD text and status is checked against the enum, so a
// corrupt row fails loudly rather than yielding a bad Listing.
func scanListing(rows *sql.Rows) (domain.Listing, error) {
	var (
		l              domain.Listing
		listed, status string
	)
	if err := rows.Scan(&l.ID, &l.Source, &l.Address, &l.City, &l.State, &l.Zip,
		&l.Price, &l.Bedrooms, &l.Bathrooms, &l.Sqft, &l.Latitude, &l.Longitude,
		&listed, &status, &l.Description); err != nil {
		return domain.Listing{}, err
	}
	date, err := time.Parse(domain.DateLayout, listed)
	if err != nil {
		return domain.Listing{}, fmt.Errorf("listing %s/%s: listed_date: %w", l.Source, l.ID, err)
	}
	l.ListedDate = date
	if l.Status, err = domain.ParseStatus(status); err != nil {
		return domain.Listing{}, fmt.Errorf("listing %s/%s: status: %w", l.Source, l.ID, err)
	}
	return l, nil
}

// Cities returns the distinct non-empty city names, compared ignoring ASCII
// case, sorted A to Z, each in the casing of its first-inserted listing
// (docs/SPEC.md section 5.2). It never returns a nil slice.
func (s *Store) Cities(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, selectCities)
	if err != nil {
		return nil, fmt.Errorf("cities: %w", err)
	}
	defer func() { _ = rows.Close() }()
	cities := []string{}
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			return nil, fmt.Errorf("cities: %w", err)
		}
		cities = append(cities, city)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cities: %w", err)
	}
	return cities, nil
}
