package trace

import (
	"context"

	"github.com/fighterlyt/common/helpers"
	"github.com/gin-gonic/gin"
)

type F11[T any, V any] struct {
}

func (o F11[T, V]) ExecFromGin(fun func(t T) (v V), t T, c *gin.Context, msg string) (v V) {
	span := helpers.StartChild(helpers.GetSpanFromGin(c), msg)
	v = fun(t)

	span.FinishSpan(nil)

	return v
}

func (o F11[T, V]) ExecFromCtx(ctx context.Context, fun func(t T) (v V), t T, msg string) (v V) {
	span := helpers.StartChild(helpers.GetSpanFromCtx(ctx), msg)
	v = fun(t)

	span.FinishSpan(nil)

	return v
}

type F11E[T any, V any] struct {
}

func (o F11E[T, V]) Gin(fun func(t T) (v V, err error), t T, c *gin.Context, msg string) (v V, err error) {
	span := helpers.StartChild(helpers.GetSpanFromGin(c), msg)
	v, err = fun(t)

	span.FinishSpan(err)

	return v, err
}

func (o F11E[T, V]) Ctx(ctx context.Context, fun func(t T) (v V, err error), t T, msg string) (v V, err error) {
	span := helpers.StartChild(helpers.GetSpanFromCtx(ctx), msg)
	v, err = fun(t)

	span.FinishSpan(err)

	return v, err
}

type F21[IN1 any, IN2 any, OUT any] struct {
}

func (t F21[IN1, IN2, OUT]) Gin(fun func(in1 IN1, in2 IN2) (out OUT), in1 IN1, in2 IN2, c *gin.Context, msg string) (out OUT) { //nolint:lll,,revive
	span := helpers.StartChild(helpers.GetSpanFromGin(c), msg)
	out = fun(in1, in2)

	span.FinishSpan(nil)

	return out
}

func (t F21[IT1, IT2, OUT]) Ctx(ctx context.Context, fun func(in1 IT1, in2 IT2) (out OUT), in1 IT1, in2 IT2, msg string) (out OUT) { //nolint:lll,,revive
	span := helpers.StartChild(helpers.GetSpanFromCtx(ctx), msg)
	out = fun(in1, in2)

	span.FinishSpan(nil)

	return out
}

type F21E[IN1 any, IN2 any, OUT any] struct {
}

func (t F21E[IT1, IT2, OUT]) Gin(fun func(in1 IT1, in2 IT2) (out OUT, err error), in1 IT1, in2 IT2, c *gin.Context, msg string) (out OUT, err error) { //nolint:lll,,revive
	span := helpers.StartChild(helpers.GetSpanFromGin(c), msg)
	out, err = fun(in1, in2)

	span.FinishSpan(err)

	return out, err
}

func (t F21E[IT1, IT2, OUT]) Ctx(ctx context.Context, fun func(in1 IT1, in2 IT2) (out OUT, err error), in1 IT1, in2 IT2, msg string) (out OUT, err error) { //nolint:lll,revive
	span := helpers.StartChild(helpers.GetSpanFromCtx(ctx), msg)
	out, err = fun(in1, in2)

	span.FinishSpan(err)

	return out, err
}
