package badger

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/fighterlyt/common/localdb"
	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Service 服务
type Service struct {
	db     *badger.DB
	logger log.Logger
}

/*
NewService 新建服务
参数:
*	filePath	string    	路径
*	logger  	log.Logger  日志器
返回值:
*	service 	*Service  	服务
*	err     	error     	错误
*/
func NewService(filePath string, logger log.Logger) (service *Service, err error) {
	service = &Service{
		logger: logger,
	}

	option := badger.DefaultOptions(filePath)
	option.Logger = newLogger(logger)

	if service.db, err = badger.Open(option); err != nil {
		return nil, errors.Wrap(err, `Open`)
	}

	return service, nil
}

func (s Service) Close() error {
	return s.db.Close()
}

func (s Service) Read(key []byte, data localdb.Item) error {
	s.logger.Info(`Get`, zap.ByteString(`key`, key))

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		if err != nil {
			return errors.Wrap(err, `txn.Get`)
		}

		if err = item.Value(func(val []byte) error {
			if err = data.Decode(val); err != nil {
				return errors.Wrap(err, `decode`)
			}

			return nil
		}); err != nil {
			return errors.Wrap(err, `Value`)
		}

		return nil
	})

	return err
}

func (s Service) Write(data localdb.Item) error {
	if data == nil {
		return nil
	}

	if err := s.db.Update(func(txn *badger.Txn) error {
		value, err := data.Encode()
		if err != nil {
			return errors.Wrap(err, `MarshalJSON`)
		}

		if err = txn.SetEntry(badger.NewEntry(data.Key(), value)); err != nil {
			return errors.Wrap(err, `SetEntry`)
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, `Update`)
	}

	if err := s.db.Sync(); err != nil {
		return errors.Wrap(err, `Sync`)
	}

	return nil
}

func (s Service) IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	return errors.As(err, &badger.ErrKeyNotFound)
}
func (s Service) Delete(key []byte) error {
	txn := s.db.NewTransaction(true)

	if err := txn.Delete(key); err != nil {
		return errors.Wrap(err, `Delete`)
	}

	if err := txn.Commit(); err != nil {
		return errors.Wrap(err, `Commit`)
	}

	return nil
}
