package storage

import "fmt"

var (
	ErrPersonNotFound  = fmt.Errorf("person not found")
	ErrNoUpdatedFields = fmt.Errorf("no updated fields")
)
