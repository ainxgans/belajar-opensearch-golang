package export

import (
	"io"
	"strings"

	"belajar-opensearch-golang/internal/product"

	"github.com/xuri/excelize/v2"
)

func WriteXLSX(w io.Writer, items []product.Product) error {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Products"
	idx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	f.SetActiveSheet(idx)
	f.DeleteSheet("Sheet1")

	headers := []string{"SKU", "Name", "Brand", "Category", "Price", "Stock", "Rating", "Tags"}
	for c, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for r, p := range items {
		row := r + 2
		set := func(col int, v any) {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			f.SetCellValue(sheet, cell, v)
		}
		set(1, p.SKU)
		set(2, p.Name)
		set(3, p.Brand)
		set(4, p.Category)
		set(5, p.Price)
		set(6, p.Stock)
		set(7, p.Rating)
		set(8, strings.Join(p.Tags, ", "))
	}
	return f.Write(w)
}
