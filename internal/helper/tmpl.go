package helper

import "html/template"

// TmplNumIterateFunc is custom func for go template to loop by a given number
// I use for rendering the dropdown of page nubmers
func TmplNumIterateFunc() template.FuncMap {
	return template.FuncMap{
		"Iterate": func(totalPages *int64) []int64 {
			var i int64
			var page []int64
			for i = 0; i < (*totalPages); i++ {
				page = append(page, i+1)
			}
			return page
		},
	}
}
