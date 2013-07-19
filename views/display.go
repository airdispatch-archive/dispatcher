package views

import (
	"airdispat.ch/airdispatch"
	library "github.com/airdispatch/go-pressure"
	"html/template"
	"time"
	"bytes"
	"fmt"
)

type TemplateTag func(interface{}) interface{}
type TemplateTagWithInteger func(interface{}, int) interface{}

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

func DisplayAirDispatchField() TemplateTagWithInteger {
	return func(arg interface{}, counter int) interface{} {
		argMap := arg.(map[string]interface{})
		output := template.HTML("")

		if argMap["Editor"] == DispatcherTextEditor {
			output = template.HTML(fmt.Sprintf("<input type='text' id='%v' class='span5' name='content[%v][1]' value='%v'>", argMap["TypeName"], counter, argMap["Payload"]))
		} else if argMap["Editor"] == DispatcherTextArea {
			output = template.HTML(fmt.Sprintf("<textarea id='%v' class='span5' name='content[%v][1]' style='height: 250px;'>%v</textarea>", argMap["TypeName"], counter, argMap["Payload"]))
		}

		return output
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

		allContent, tmpName := DetectMessageType(UnmarshalMessagePayload(mail))
		context["Content"] = allContent

		context["GetContent"] = GetContent(allContent)
		context["DisplayAddress"] = DisplayAirDispatchAddress(s)

		theBuffer := bytes.NewBuffer(nil)
		s.WriteTemplateToBuffer(tmpName, "generic", theBuffer, context)

		return template.HTML(theBuffer.String())
	}
}

func DetectMessageType(arg []*airdispatch.MailData_DataType) (map[string]interface{}, string) {
	return GetNamedMapFromPayload(arg, func(data []byte)interface{} {
			return string(data)
	}), "display/blog.html"
}

type ADUnloader func([]byte) interface{}

func GetNamedMapFromPayload(content []*airdispatch.MailData_DataType, unload ADUnloader) map[string]interface{} {
	allContent := make(map[string]interface{})

	for _, v := range(content) {
		allContent[*v.TypeName] = unload(v.Payload)
	}

	return allContent
}