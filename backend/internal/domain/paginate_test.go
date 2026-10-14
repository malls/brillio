package domain_test

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"brillio/backend/internal/domain"
)

// scoredN returns n ranked listings with IDs "L01", "L02", ... in rank order.
func scoredN(n int) []domain.Scored {
	out := make([]domain.Scored, n)
	for i := range out {
		out[i] = domain.Scored{
			Listing:   domain.Listing{Source: "MLS_A", ID: fmt.Sprintf("L%02d", i+1)},
			Relevance: domain.Relevance{Score: 1 - float64(i)/float64(n), Recency: 1},
		}
	}
	return out
}

func TestPaginate(t *testing.T) {
	tests := []struct {
		name           string
		n              int // ranked items
		page, pageSize int
		wantLen        int
		wantFirstRank  int // rank of the first item; unused when wantLen is 0
		wantTotalPages int
	}{
		{name: "page_1", n: 25, page: 1, pageSize: 10, wantLen: 10, wantFirstRank: 1, wantTotalPages: 3},
		{name: "last_partial_page", n: 25, page: 3, pageSize: 10, wantLen: 5, wantFirstRank: 21, wantTotalPages: 3},
		{name: "last_full_page", n: 20, page: 2, pageSize: 10, wantLen: 10, wantFirstRank: 11, wantTotalPages: 2},
		{name: "page_past_end", n: 25, page: 4, pageSize: 10, wantLen: 0, wantTotalPages: 3},
		{name: "page_past_end_max_int", n: 25, page: math.MaxInt, pageSize: 100, wantLen: 0, wantTotalPages: 1},
		{name: "page_size_1", n: 25, page: 25, pageSize: 1, wantLen: 1, wantFirstRank: 25, wantTotalPages: 25},
		{name: "page_size_100", n: 150, page: 1, pageSize: 100, wantLen: 100, wantFirstRank: 1, wantTotalPages: 2},
		{name: "page_size_100_last_page", n: 150, page: 2, pageSize: 100, wantLen: 50, wantFirstRank: 101, wantTotalPages: 2},
		{name: "zero_items", n: 0, page: 1, pageSize: 10, wantLen: 0, wantTotalPages: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ranked []domain.Scored // nil for zero_items
			if tt.n > 0 {
				ranked = scoredN(tt.n)
			}
			got, err := domain.Paginate(ranked, tt.page, tt.pageSize)
			if err != nil {
				t.Fatalf("Paginate(%d items, %d, %d) error = %v, want nil", tt.n, tt.page, tt.pageSize, err)
			}
			if got.Page != tt.page || got.PageSize != tt.pageSize || got.TotalItems != tt.n || got.TotalPages != tt.wantTotalPages {
				t.Errorf("metadata = {Page:%d PageSize:%d TotalItems:%d TotalPages:%d}, want {Page:%d PageSize:%d TotalItems:%d TotalPages:%d}",
					got.Page, got.PageSize, got.TotalItems, got.TotalPages, tt.page, tt.pageSize, tt.n, tt.wantTotalPages)
			}
			if got.Items == nil {
				t.Fatalf("Items is nil, want a non-nil slice")
			}
			if len(got.Items) != tt.wantLen {
				t.Fatalf("len(Items) = %d, want %d", len(got.Items), tt.wantLen)
			}
			for k, it := range got.Items {
				if want := tt.wantFirstRank + k; it.Rank != want {
					t.Errorf("Items[%d].Rank = %d, want %d", k, it.Rank, want)
					continue
				}
				if want := ranked[it.Rank-1].ID; it.ID != want {
					t.Errorf("Items[%d] (rank %d) ID = %q, want %q", k, it.Rank, it.ID, want)
				}
			}
		})
	}

	t.Run("rank_continuity_across_pages", func(t *testing.T) {
		ranked := scoredN(25)
		for _, size := range []int{1, 7, 10, 100} {
			first, err := domain.Paginate(ranked, 1, size)
			if err != nil {
				t.Fatalf("size %d: Paginate error = %v", size, err)
			}
			var ranks []int
			var ids []string
			for p := 1; p <= first.TotalPages; p++ {
				got, err := domain.Paginate(ranked, p, size)
				if err != nil {
					t.Fatalf("size %d page %d: Paginate error = %v", size, p, err)
				}
				for _, it := range got.Items {
					ranks = append(ranks, it.Rank)
					ids = append(ids, it.ID)
				}
			}
			if len(ranks) != len(ranked) {
				t.Errorf("size %d: walked %d items over %d pages, want %d", size, len(ranks), first.TotalPages, len(ranked))
				continue
			}
			for i := range ranks {
				if ranks[i] != i+1 || ids[i] != ranked[i].ID {
					t.Errorf("size %d: item %d = rank %d %s, want rank %d %s", size, i, ranks[i], ids[i], i+1, ranked[i].ID)
				}
			}
		}

		page2, err := domain.Paginate(ranked, 2, 10)
		if err != nil {
			t.Fatalf("Paginate(page 2, size 10) error = %v", err)
		}
		if len(page2.Items) == 0 || page2.Items[0].Rank != 11 || page2.Items[0].ID != "L11" {
			t.Errorf("page 2 of size 10 starts with %+v, want rank 11 ID L11", page2.Items[:min(1, len(page2.Items))])
		}
	})
}

func TestPaginateInvalid(t *testing.T) {
	tests := []struct {
		name           string
		page, pageSize int
		wantPageErr    bool
		wantSizeErr    bool
	}{
		{name: "invalid_page_0", page: 0, pageSize: 10, wantPageErr: true},
		{name: "invalid_page_negative", page: -1, pageSize: 10, wantPageErr: true},
		{name: "invalid_page_size_0", page: 1, pageSize: 0, wantSizeErr: true},
		{name: "invalid_page_size_101", page: 1, pageSize: 101, wantSizeErr: true},
		{name: "invalid_page_size_negative", page: 1, pageSize: -5, wantSizeErr: true},
		{name: "invalid_page_and_page_size", page: 0, pageSize: 0, wantPageErr: true, wantSizeErr: true},
	}
	ranked := scoredN(25)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.Paginate(ranked, tt.page, tt.pageSize)
			if err == nil {
				t.Fatalf("Paginate(%d, %d) error = nil, want an error", tt.page, tt.pageSize)
			}
			if is := errors.Is(err, domain.ErrPageOutOfRange); is != tt.wantPageErr {
				t.Errorf("errors.Is(%v, ErrPageOutOfRange) = %v, want %v", err, is, tt.wantPageErr)
			}
			if is := errors.Is(err, domain.ErrPageSizeOutOfRange); is != tt.wantSizeErr {
				t.Errorf("errors.Is(%v, ErrPageSizeOutOfRange) = %v, want %v", err, is, tt.wantSizeErr)
			}
			if !reflect.DeepEqual(got, domain.Page{}) {
				t.Errorf("Paginate(%d, %d) page = %+v, want the zero Page", tt.page, tt.pageSize, got)
			}
		})
	}
}
