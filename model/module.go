package model

type Module interface {
	Close()
	IsClosed() bool
	Key() string
	Name() string
}
