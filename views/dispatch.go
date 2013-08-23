package views

import (
	"airdispat.ch/client/framework"
	"airdispat.ch/common"
	"github.com/airdispatch/dispatcher/library"
	"github.com/airdispatch/dispatcher/models"
)

type UserProfile struct {
	Name  string
	Image string
}

func GetProfile(user string, key *common.ADKey, s library.Server) (UserProfile, error) {
	c := framework.Client{}
	c.Populate(key)

	location, err := common.LookupLocation(user, models.GetTrackerList(s.DbMap), key)
	if err != nil {
		return nil, err
	}

	mail, err := c.DownloadSpecificMessageFromServer((user + "::profile"), "")
	if err != nil {
		return nil, err
	}

	internalData := UnmarshalMessagePayload(mail)

	outputProfile := UserProfile{}

	for i, v := range internalData {
		if v.GetTypeName() == "airdispat.ch/profiles/name" {
			outputProfile.Name = string(v.GetPayload())
		} else if v.GetTypeName() == "airdispat.ch/profiles/image" {
			outputProfile.Image = string(v.GetPayload())
		}
	}

	return outputProfile
}
