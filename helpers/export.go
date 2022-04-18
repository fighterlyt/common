package helpers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

type ExportRecord interface {
	GetExportFields() []interface{}
}

func BuildXLSFile(writer io.Writer, headers map[string]int, headerOrders []string, title string, records ...ExportRecord) error {
	for i, record := range records {
		if len(headers) != len(record.GetExportFields()) {
			return fmt.Errorf(`第[%d]条记录,字段数量错误`, i+1)
		}
	}

	f, err := CreateXLSFile(headers, headerOrders, title, records...)
	if err != nil {
		return err
	}

	if err = f.Write(writer); err != nil {
		return errors.Wrap(err, `写入应答`)
	}

	return nil
}

// 专门处理导出为xlsx
func BuildXLSX(ctx *gin.Context, headers map[string]int, headerOrders []string, fileName, title string, records ...ExportRecord) error {
	var b bytes.Buffer

	if err := BuildXLSFile(&b, headers, headerOrders, title, records...); err != nil {
		return err
	}

	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header(`Content-Disposition`, fmt.Sprintf(`attachment;filename=%s`, fileName))
	ctx.Data(http.StatusOK, "application/octet-stream", b.Bytes())

	return nil
}

func charIncr(from byte) string {
	return string(from + 1)
}

func CharAdd(from byte, delta int) byte {
	for ; delta > 0; delta-- {
		from = charIncr(from)[0]
	}

	return from
}

// CreateXLSFile 创建xlsx文件
func CreateXLSFile(headers map[string]int, headerOrders []string, title string, records ...ExportRecord) (*excelize.File, error) {
	f := excelize.NewFile()

	sheet1 := `Sheet1`
	// Create a new sheet.
	index := f.NewSheet(sheet1)
	style, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{
		Horizontal: "center",
	}})

	last := CharAdd('A', len(headers)-1)
	f.MergeCell(sheet1, `A1`, string(last)+`1`)

	f.SetCellValue(sheet1, `A1`, title)

	i := 0
	for _, text := range headerOrders {

		column := string(CharAdd('A', i))
		f.SetColStyle(sheet1, column, style)
		f.SetColWidth(sheet1, column, column, float64(headers[text]))

		f.SetCellValue(sheet1, column+"2", text)
		i++
	}

	// Set active sheet of the workbook.
	f.SetActiveSheet(index)
	// Save spreadsheet by the given path.

	for i, record := range records {
		for j, field := range record.GetExportFields() {
			column := string(CharAdd('A', j))
			f.SetCellValue(sheet1, fmt.Sprintf(`%s%d`, column, i+3), field)
		}
	}

	return f, nil
}
