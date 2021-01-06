package secKill

import "context"

type ContextStruct struct {
	Tag string
	Ctx context.Context
	Cancel context.CancelFunc
}

func NewContextStruct(c context.Context, cc context.CancelFunc, tag string) *ContextStruct {
	return &ContextStruct{
		Tag:    tag,
		Ctx:    c,
		Cancel: cc,
	}
}