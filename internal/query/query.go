package query

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func Filter(items []map[string]interface{}, q url.Values) []map[string]interface{} {
	var out []map[string]interface{}
	for _, item := range items {
		match := true
		for key, vals := range q {
			if strings.HasPrefix(key, "_") {
				continue
			}
			if len(vals) > 0 && fmt.Sprintf("%v", item[key]) != vals[0] {
				match = false
				break
			}
		}
		if search := q.Get("_search"); search != "" {
			found := false
			for _, v := range item {
				if str, ok := v.(string); ok && strings.Contains(strings.ToLower(str), strings.ToLower(search)) {
					found = true
					break
				}
			}
			if !found {
				match = false
			}
		}
		if match {
			out = append(out, item)
		}
	}
	return out
}

func SortItems(items []map[string]interface{}, q url.Values, defaultField, defaultOrder string) []map[string]interface{} {
	field := q.Get("_sort")
	if field == "" {
		field = defaultField
	}
	if field == "" {
		return items
	}
	order := q.Get("_order")
	if order == "" {
		order = defaultOrder
	}
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i][field], items[j][field]
		ai, aok := a.(float64)
		bi, bok := b.(float64)
		if aok && bok {
			if order == "desc" {
				return ai > bi
			}
			return ai < bi
		}
		sa := fmt.Sprintf("%v", a)
		sb := fmt.Sprintf("%v", b)
		if order == "desc" {
			return sa > sb
		}
		return sa < sb
	})
	return items
}

func Paginate(items []map[string]interface{}, page, limit int) ([]map[string]interface{}, int) {
	total := len(items)
	if page < 1 {
		page = 1
	}
	start := (page - 1) * limit
	if start >= total {
		return []map[string]interface{}{}, total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total
}

func ParsePageLimit(q url.Values, defaultPage, defaultLimit int) (int, int) {
	page := defaultPage
	limit := defaultLimit
	if p := q.Get("_page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := q.Get("_limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	return page, limit
}
