package service

import "strings"

// Wrapps the isOwner logic
func isOwnerWrapper(args ...interface{}) (interface{}, error) {
	reqSub := args[0].(string)
	reqObj := args[1].(string)
	polObj := args[2].(string)

	return bool(isOwner(reqSub, reqObj, polObj)), nil
}

func isOwner(reqSub string, reqObj string, polObj string) bool {
	// First find out if there is a UID
	if strings.Contains(polObj, "{uid}") {
		comparable := strings.Replace(polObj, "{uid}", reqSub, 1)

		if comparable == reqObj {
			return true
		}
	}

	return false
}
