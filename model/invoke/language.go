package invoke

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var (
	language = map[string]map[string]string{}
)

func loadLanguage(path string) error {
	var (
		file *os.File
	)

	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		println(path, filepath.Ext(path), filepath.Base(path))
		if filepath.Ext(path) != `.json` {
			return nil
		}

		lang := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		temp := make(map[string]string, 100)

		if file, err = os.Open(path); err != nil {
			return errors.Wrapf(err, `读取文件%s`, path)
		}

		if err = json.NewDecoder(file).Decode(&temp); err != nil {
			return errors.Wrapf(err, `解析文件%s`, path)
		}
		language[lang] = temp
		return nil
	}); err != nil {
		return errors.Wrapf(err, `加载`)
	}

	return nil
}

// translateJSON contains the given interface object.
type translateJSON struct {
	Data interface{}
	lang string
}

// Render (translateJSON) writes data with custom ContentType.
func (r translateJSON) Render(w http.ResponseWriter) (err error) {
	if err = WriteJSON(w, r.Data, r.lang); err != nil {
		panic(err)
	}
	return
}

// WriteContentType (translateJSON) writes translateJSON ContentType.
func (r translateJSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// WriteJSON marshals the given interface object and writes it with custom ContentType.
func WriteJSON(w http.ResponseWriter, obj interface{}, lang string) error {
	writeContentType(w, jsonContentType)
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	jsonBytes = translate(language, lang, jsonBytes)
	_, err = w.Write(jsonBytes)
	return err
}

func translate(langs map[string]map[string]string, lang string, data []byte) []byte {
	if langs == nil {
		return data
	}

	switch lang {
	case ``:
		return data
	default:
		dict := langs[lang]

		if dict == nil {
			return data
		}

		for candidate, translate := range dict {
			if bytes.Contains(data, []byte(candidate)) {
				data = bytes.ReplaceAll(data, []byte(candidate), []byte(translate))
			}
		}

		return data
	}
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

var jsonContentType = []string{"application/json; charset=utf-8"}
