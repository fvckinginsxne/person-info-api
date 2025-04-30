package dto

import "person-info/internal/domain/model"

type PersonResponse struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic"`
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"nationality"`
}

func ToPersonResponse(p *model.Person) *PersonResponse {
	return &PersonResponse{
		Name:        p.Name,
		Surname:     p.Surname,
		Patronymic:  p.Patronymic,
		Age:         p.Age,
		Gender:      p.Gender,
		Nationality: p.Nationality,
	}
}
