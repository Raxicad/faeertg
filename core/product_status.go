package core

type ProductStatus int

const (
	Available ProductStatus = iota
	Unavailable
)

type ProductRef struct {
	url    string
	status ProductStatus
}
