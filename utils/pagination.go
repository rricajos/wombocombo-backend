package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int64 `json:"total"`
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func ParsePagination(c *fiber.Ctx) Pagination {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return Pagination{Page: page, PerPage: perPage}
}
