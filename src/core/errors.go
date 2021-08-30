package core

type ApplicationError string

func (a ApplicationError) Error() string {
	return string(a)
}
