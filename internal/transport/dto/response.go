package dto

import "person-info/internal/domain/model"

type ErrorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}

type PersonResponse struct {
	Name        string `json:"name" example:"Matvey"`
	Surname     string `json:"surname" example:"Likhanov"`
	Patronymic  string `json:"patronymic" example:"Dmitrievich"`
	Age         int    `json:"age" example:"20"`
	Gender      string `json:"gender" example:"Male"`
	Nationality string `json:"nationality" example:"RU"`
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
