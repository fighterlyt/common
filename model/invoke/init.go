package invoke

import "github.com/pkg/errors"

func Init(path string) error {
	if err := loadLanguage(path); err != nil {
		return errors.Wrap(err, `加载多语言`)
	}

	return nil
}
