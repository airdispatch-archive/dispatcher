package views

import (
	"dispatcher/library"
	"time"
)

type TemplateTag func(interface{}) interface{}

func TimestampToString() TemplateTag {
	return func(arg interface{}) interface{} {
		timestamp := arg.(int64)
		return time.Unix(timestamp, 0).Format("Jan 2, 2006 at 3:04pm")
	}
}

func DisplayAirDispatchAddress(s *library.Server) TemplateTag {
	return func(arg interface{}) interface{} {
		return arg
	}
}

func DisplayMessageTag() TemplateTag {
	return func(arg interface{}) interface{} {
		return arg
	}
}