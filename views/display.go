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

func GetContent(allContent map[string]interface{}) TemplateTag {
	return func(arg interface{}) interface{} {
		stringArg := arg.(string)
		return allContent[stringArg]
	}
}

func DisplayMessageTag(s *library.Server) TemplateTag {
	return func(arg interface{}) interface{} {
		context := make(map[string]interface{})

		mail := arg.(*airdispatch.Mail)

		context["FromAddress"] = mail.FromAddress
		context["Encryption"] = mail.Encryption

		allContent := DetectMessageType(UnmarshalMessagePayload(mail))
		context["Content"] = allContent

		context["GetContent"] = GetContent(allContent)

		context["DisplayAddress"] = DisplayAirDispatchAddress(s)

		theBuffer := bytes.NewBuffer(nil)
		s.WriteTemplateToBuffer("display/blog.html", "generic", theBuffer, context)

		return template.HTML(theBuffer.String())
	}
}

func DetectMessageType(arg []*airdispatch.MailData_DataType) map[string]interface{} {
	return GetNamedMapFromPayload(arg, func(data []byte)interface{} {
			return string(data)
	})
}

type ADUnloader func([]byte) interface{}

func GetNamedMapFromPayload(content []*airdispatch.MailData_DataType, unload ADUnloader) map[string]interface{} {
	allContent := make(map[string]interface{})

	for _, v := range(content) {
		allContent[*v.TypeName] = unload(v.Payload)
	}

	return allContent
}