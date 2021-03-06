package parse

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/rancher/norman/types"
)

var (
	defaultLimit = int64(100)
	maxLimit     = int64(3000)
)

func QueryOptions(req *http.Request, schema *types.Schema) *types.QueryOptions {
	if req.Method != http.MethodGet {
		return nil
	}

	result := &types.QueryOptions{}

	result.Sort = parseSort(schema, req)
	result.Pagination = parsePagination(req)
	result.Conditions = parseFilters(schema, req)

	return result
}

func parseOrder(req *http.Request) types.SortOrder {
	order := req.URL.Query().Get("order")
	if types.SortOrder(order) == types.DESC {
		return types.DESC
	}
	return types.ASC
}

func parseSort(schema *types.Schema, req *http.Request) types.Sort {
	sortField := req.URL.Query().Get("sort")
	if _, ok := schema.CollectionFilters[sortField]; !ok {
		sortField = ""
	}
	return types.Sort{
		Order: parseOrder(req),
		Name:  sortField,
	}
}

func parsePagination(req *http.Request) *types.Pagination {
	q := req.URL.Query()
	limit := q.Get("limit")
	marker := q.Get("marker")

	result := &types.Pagination{
		Limit:  &defaultLimit,
		Marker: marker,
	}

	if limit != "" {
		limitInt, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return result
		}

		if limitInt > maxLimit {
			result.Limit = &maxLimit
		} else if limitInt > 0 {
			result.Limit = &limitInt
		}
	}

	return result
}

func parseNameAndOp(value string) (string, string) {
	name := value
	op := "eq"

	idx := strings.LastIndex(value, "_")
	if idx > 0 {
		op = value[idx+1:]
		name = value[0:idx]
	}

	return name, op
}

func parseFilters(schema *types.Schema, req *http.Request) []*types.QueryCondition {
	conditions := []*types.QueryCondition{}
	for key, values := range req.URL.Query() {
		name, op := parseNameAndOp(key)
		filter, ok := schema.CollectionFilters[name]
		if !ok {
			continue
		}

		for _, mod := range filter.Modifiers {
			if op != mod || !types.ValidMod(op) {
				continue
			}

			genericValues := []interface{}{}
			for _, value := range values {
				genericValues = append(genericValues, value)
			}

			conditions = append(conditions, types.NewConditionFromString(name, mod, genericValues))
		}
	}

	return conditions
}
