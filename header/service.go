package header

import (
	"github.com/fighterlyt/common/helpers"
)

type service struct {
	*helpers.BaseResource
}

func NewService(resource *helpers.BaseResource) (target *service, err error) {
	target = &service{resource}
	target.http()

	return target, nil
}

func (s *service) Register(items Items) error {
	return Register(items)
}
