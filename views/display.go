package views

import (
	"airdispat.ch/airdispatch"
	"github.com/airdispatch/dispatcher/library"
	"html/template"
	"time"
	"bytes"
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

func DisplayMessageTag(s *library.Server) TemplateTag {
	return func(arg interface{}) interface{} {
		context := make(map[string]interface{})

		mail := arg.(*airdispatch.Mail)

		context["FromAddress"] = mail.FromAddress
		context["Encryption"] = mail.Encryption

		allContent := GetNamedMapFromPayload(UnmarshalMessagePayload(mail))
		context["Content"] = allContent

		context["GetContent"] = func(contentArg string) interface{} {
			return allContent[contentArg]
		}

		context["DisplayAddress"] = DisplayAirDispatchAddress(s)

		theBuffer := bytes.NewBuffer(nil)
		s.WriteTemplateToBuffer("display/blog.html", "generic", theBuffer, context)

		return template.HTML(theBuffer.String())
	}
}