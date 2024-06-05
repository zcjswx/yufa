package app

type UnauthError struct {
}

func (UnauthError) Error() string {
	return "Request Unauthorized"
}
