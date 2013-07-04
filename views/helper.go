package views

import (
	"github.com/airdispatch/dispatcher/models"
	"github.com/airdispatch/dispatcher/library"
	"github.com/hoisie/web"
	"code.google.com/p/goprotobuf/proto"
	"airdispat.ch/airdispatch"
	"errors"
	"strconv"
)

const noEncryption = "none"

func GetLoggedInUser(s *library.Server, ctx *web.Context) (*models.User) {
	session, err := s.GetMainSession(ctx)

	if session.Values[LoginSessionMapKey] == nil || session.Values[LoginSessionMapKey] == "" || session.Values[LoginSessionMapKey] == -1 {
		return nil
	}

	user, err := s.DbMap.Get(models.User{}, session.Values[LoginSessionMapKey])
	if err != nil {
		return nil
	}

	if user == nil {
		return nil
	}

	newUser := user.(*models.User)
	newUser.Populate()

	return newUser
}

func WildcardTemplateLoginRequired(s *library.Server, t library.WildcardTemplateView) library.WildcardTemplateView {
	return func(ctx *web.Context, val string) {
		u := GetLoggedInUser(s, ctx)
		if (u != nil) {
			t(ctx, val)
		} else {
			ctx.Redirect(303, "/login")
		}
	}
}

func TemplateLoginRequired(s *library.Server, t library.TemplateView) library.TemplateView {
	return func(ctx *web.Context) {
		u := GetLoggedInUser(s, ctx)
		if (u != nil) {
			t(ctx)
		} else {
			ctx.Redirect(303, "/login")
		}
	}
}

func LogoutView(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		session, _ := s.GetMainSession(ctx)
		session.Values[LoginSessionMapKey] = -1
		library.SaveSessionWithContext(session, ctx)
		ctx.Redirect(303, "/login")
	}
}

func MessageToContext(m *models.Message, s *library.Server) map[string]interface{} {
	output := make(map[string]interface{})

	output["ID"] = m.Id
	output["MesType"] = m.MessageType

	if m.ToAddress == "" {
		output["TO"] = "Public"
		output["ToAddress"] = ""
		output["Encryption"] = ""
	} else {
		output["TO"] = DisplayAirDispatchAddress(s)(m.ToAddress)
		output["ToAddress"] = output["TO"]
		output["Encryption"] = "aes/256"
	}

	theUser, _ := s.DbMap.Get(&models.User{}, m.SendingUser)
	output["FROM"] = theUser.(*models.User).FullName

	output["Timestamp"] = TimestampToString()(m.Timestamp)

	theData := &airdispatch.MailData{}
	proto.Unmarshal(m.Content, theData)

	output["Content"] = GetContextFromPayload(theData.Payload)

	return output
}

func UnmarshalMessagePayload(message *airdispatch.Mail) []*airdispatch.MailData_DataType {
	if *message.Encryption == noEncryption {
		theContent := &airdispatch.MailData{}
		proto.Unmarshal(message.Data, theContent)

		return theContent.Payload
	}
	return nil
}

func GetNamedMapFromPayload(content []*airdispatch.MailData_DataType) map[string]interface{} {
	allContent := make(map[string]interface{})

	for _, v := range(content) {
		allContent[*v.TypeName] = v.Payload
	}

	return allContent
}

func GetContextFromPayload(content []*airdispatch.MailData_DataType) []map[string]interface{} {
	allContent := make([]map[string]interface{}, len(content))

	for i, v := range(content) {
		allContent[i] = map[string]interface{} {
			"TypeName": v.TypeName,
			"Payload": string(v.Payload),
		}

		// TODO: Change the Field Type Here...
	}

	return allContent
}

func ContextToDataTypeBytes(ctx *web.Context) ([]byte, error) {
	if len(ctx.Params) <= 2 {
		return nil, errors.New("The Context doesn't have enough Fields")
	}

	content_types := make([]*airdispatch.MailData_DataType, (len(ctx.Params) - 1) / 2)

	for key, value := range(ctx.Params) {
		if key != "to_address" {
			// content[index][type]
			i := contentRegex.FindAllStringSubmatch(key, -1)[0]

			typeIndex, err := strconv.ParseInt(i[1], 10, 0)
			if err != nil {
				return nil, errors.New("The Context doesn't have correct field formats.")
			}

			theData := content_types[typeIndex]
			if theData == nil {
				theData = &airdispatch.MailData_DataType {}
			}

			if i[2] == "0" {
				theType := ctx.Params[key]
				theData.TypeName = &theType
			} else if i[2] == "1" {
				theData.Payload = []byte(value)
			}

			content_types[typeIndex] = theData
		}
	}

	theData := &airdispatch.MailData{
		Payload: content_types,
	}

	byteData, err := proto.Marshal(theData)
	if err != nil {
		return nil, err
	}

	return byteData, nil
}