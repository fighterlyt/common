package free

import (
	"fmt"
	"sync"

	"github.com/fighterlyt/log"
	"go.uber.org/zap"
)

type hooks struct {
	entries map[string]Hook
	*sync.RWMutex
	logger log.Logger
}

func NewHooks(logger log.Logger) *hooks {
	return &hooks{
		entries: make(map[string]Hook, 10),
		RWMutex: &sync.RWMutex{},
		logger:  logger,
	}
}

func (h hooks) Add(hook Hook) error {
	h.Lock()
	defer h.Unlock()

	key := hook.Key()

	if _, exist := h.entries[key]; exist {
		return fmt.Errorf(`key[%s]已存在`, key)
	}

	h.entries[key] = hook

	return nil
}

func (h hooks) Remove(key string) {
	h.Lock()
	defer h.Unlock()

	delete(h.entries, key)
}

func (h hooks) EveryBeforeFreeze(info *FreezeInfo) {
	h.RLock()
	defer h.RUnlock()

	for key, hook := range h.entries {
		h.logger.Info(`beforeFreeze`, zap.String(`hook`, key))

		hook.BeforeFreeze(info)
	}
}

func (h hooks) EveryAfterFreeze(info *FreezeInfo, err error) {
	h.RLock()
	defer h.RUnlock()

	for key, hook := range h.entries {
		h.logger.Info(`afterFreeze`, zap.String(`hook`, key))

		hook.AfterFreeze(info, err)
	}
}

func (h hooks) EveryBeforeUnfreeze(info *FreezeInfo) {
	h.RLock()
	defer h.RUnlock()

	for key, hook := range h.entries {
		h.logger.Info(`beforeUnfreeze`, zap.String(`hook`, key))

		hook.BeforeUnfreeze(info)
	}
}

func (h hooks) EveryAfterUnfreeze(info *FreezeInfo, err error) {
	h.RLock()
	defer h.RUnlock()

	for key, hook := range h.entries {
		h.logger.Info(`afterUnfreeze`, zap.String(`hook`, key))

		hook.AfterUnfreeze(info, err)
	}
}
