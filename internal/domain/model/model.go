package model

type Person struct {
	Name        string
	Surname     string
	Patronymic  string
	Age         int
	Gender      string
	Nationality string
}

type PeopleFilters struct {
	Name        string
	Surname     string
	Patronymic  string
	Age         int
	Gender      string
	Nationality string
}

type Pagination struct {
	Page int
	Size int
}

type SortOptions struct {
	By    string
	Order string
}
