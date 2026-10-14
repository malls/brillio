package domain

// SearchFilter is a parsed and validated set of search filters (docs/SPEC.md
// sections 5.3 and 6.1). Query parsing produces it, the store applies it, and
// the response echoes it back as appliedFilters.
type SearchFilter struct {
	// MinPrice is the inclusive lower price bound; nil means absent.
	MinPrice *float64
	// MaxPrice is the inclusive upper price bound; nil means absent.
	MaxPrice *float64
	// MinBedrooms is the inclusive minimum bedroom count; nil means absent.
	MinBedrooms *int
	// City is the trimmed city name, matched case-insensitively and exactly;
	// "" means absent.
	City string
	// Keyword is the keyword parameter as received, echoed in appliedFilters;
	// "" means absent.
	Keyword string
	// KeywordTokens are Keyword split on whitespace and lowercased. Every token
	// must appear in a listing's description for it to match.
	KeywordTokens []string
	// Statuses are the statuses to include. It is never empty after parsing:
	// an absent status parameter means AllStatuses().
	Statuses []ListingStatus
}
