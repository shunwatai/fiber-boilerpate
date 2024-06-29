package helper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

func getDefaultSortKey ()string{
	defaultSortKey := "id"
	if cfg.DbConf.Driver == "mongodb" {
		defaultSortKey = "createdAt"
	}
	return defaultSortKey
}

func parseOrderBy(rawOrderBy string) map[string]string {
	orderBy := make(map[string]string)
	splitedOrderBy := strings.Split(rawOrderBy, ".")
	// fmt.Printf("splitedOrderBy:  %+v\n", splitedOrderBy)

	orderBy["by"] = "desc"
	if len(rawOrderBy) == 0 {
		orderBy["key"] = getDefaultSortKey()
	} else if len(splitedOrderBy) != 2 {
		orderBy["key"] = strcase.ToSnake(splitedOrderBy[0])
	} else {
		orderBy["key"] = strcase.ToSnake(splitedOrderBy[0])
		orderBy["by"] = strings.ToLower(splitedOrderBy[1])
	}

	return orderBy
}

type Pagination struct {
	Page       int64             `json:"page"`       // current page
	Items      int64             `json:"items"`      // records per page
	Count      int64             `json:"count"`      // total records
	OrderBy    map[string]string `json:"orderBy"`    // orderBy
	TotalPages int64             `json:"totalPages"` // ceil(count / items)
	NextPage   string            `json:"nextPage"`   // next page url
	PrevPage   string            `json:"prevPage"`   // prev page url
}

func (p *Pagination) SetPageUrls() {
	p.setNextPageUrl()
	p.setPrevPageUrl()
}

func (p *Pagination) setNextPageUrl() {
	var nextPageUrl string
	if p.Page >= p.TotalPages {
		nextPageUrl = fmt.Sprintf("items=%d&page=%d&orderBy=%s.%s", p.Items, p.Page, p.OrderBy["key"], p.OrderBy["by"])
	} else {
		nextPageUrl = fmt.Sprintf("items=%d&page=%d&orderBy=%s.%s", p.Items, p.Page+1, p.OrderBy["key"], p.OrderBy["by"])
	}
	p.NextPage = nextPageUrl
}

func (p *Pagination) setPrevPageUrl() {
	var prevPageUrl string
	if p.Page <= 1 {
		prevPageUrl = fmt.Sprintf("items=%d&page=%d&orderBy=%s.%s", p.Items, p.Page, p.OrderBy["key"], p.OrderBy["by"])
	} else {
		prevPageUrl = fmt.Sprintf("items=%d&page=%d&orderBy=%s.%s", p.Items, p.Page-1, p.OrderBy["key"], p.OrderBy["by"])
	}
	p.PrevPage = prevPageUrl
}

func getDefaultPagination() *Pagination {
	orderBy := map[string]string{
		"key": getDefaultSortKey(),
		"by":  "desc",
	}

	return &Pagination{
		Page:       1,
		Items:      0,
		Count:      0,
		OrderBy:    orderBy,
		TotalPages: 1,
	}
}

// get the pagination struct according to the req querystring
// key: page(page number), items(number of rows per page)
func GetPagination(queries map[string]interface{}) *Pagination {
	pagination := getDefaultPagination()

	if queries["page"] != nil && queries["items"] != nil {
		pagination.Page, _ = strconv.ParseInt(queries["page"].(string), 10, 64)
		pagination.Items, _ = strconv.ParseInt(queries["items"].(string), 10, 64)
	}

	if queries["order_by"] != nil {
		pagination.OrderBy = parseOrderBy(queries["order_by"].(string))
	}

	return pagination
}

// ensure the queries only contains the keys that match with table's columns,
// clean the irrelevant keys from the queries
func SanitiseQuerystring(cols []string, queries map[string]interface{}) {
	tmpColsMap := map[string]struct{}{}
	for _, col := range cols {
		tmpColsMap[col] = struct{}{}
	}
	for k := range queries {
		_, ok := tmpColsMap[k]
		if !ok {
			delete(queries, k)
		}
	}
}
