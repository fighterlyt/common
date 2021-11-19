package helpers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBuildXLSX(t *testing.T) {
	ctx := gin.Default()

	type args struct {
		headers      map[string]int
		headerOrders []string
		fileName     string
		title        string
		records      []ExportRecord
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: `3行3列`,
			args: args{
				headers: map[string]int{
					`A`: 10,
					`B`: 5,
					`C`: 10,
				},
				headerOrders: []string{`A`, `B`, `C`},
				fileName:     "A",
				title:        "测试",
				records: []ExportRecord{mockExport{values: []interface{}{10, `a`, 1.1}},
					mockExport{
						values: []interface{}{20, `b`, 2.2},
					},
					mockExport{
						values: []interface{}{30, `c`, 3.3},
					},
				},
			},
		},
	}
	ctx.GET(`/export/:index`, func(ctx *gin.Context) {
		index, err := strconv.ParseInt(ctx.Param(`index`), 10, 64)
		require.NoError(t, err)
		tt := tests[index]
		if err := BuildXLSX(ctx, tt.args.headers, tt.args.headerOrders, tt.args.fileName, tt.args.title, tt.args.records...); (err != nil) != tt.wantErr {
			t.Errorf("BuildXLSX() error = %v, wantErr %v", err, tt.wantErr)
		}
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export/0", nil)
	ctx.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)

	file, err := os.Create(`1.xlsx`)
	require.NoError(t, err)

	io.Copy(file, w.Body)

	file.Close()

}

type mockExport struct {
	values []interface{}
}

func (m mockExport) GetExportFields() []interface{} {
	return m.values
}
