package common

import "github.com/pkg/errors"

var (
	ServiceNoSuchPrivacyLevel = errors.New("No such privacyLevel")
	ServiceMissingRequiredData = errors.New("Invalid user data")
)
