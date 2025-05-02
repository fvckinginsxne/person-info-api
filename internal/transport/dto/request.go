package dto

import "person-info/internal/domain/model"

type CreatePersonRequest struct {
	Name       string `json:"name" binding:"required" example:"John"`
	Surname    string `json:"surname" binding:"required" example:"Snow"`
	Patronymic string `json:"patronymic,omitempty" example:"Dmitrievich"`
}

type UpdatePersonRequest struct {
	Name        string `json:"name,omitempty" example:"John"`
	Surname     string `json:"surname,omitempty" example:"Snow"`
	Patronymic  string `json:"patronymic,omitempty" example:"Dmitrievich"`
	Age         int    `json:"age,omitempty" example:"30"`
	Gender      string `json:"gender,omitempty" example:"Male"`
	Nationality string `json:"nationality,omitempty" example:"RU"`
}

func ToPersonModel(p *UpdatePersonRequest) *model.Person {
	return &model.Person{
		Name:        p.Name,
		Surname:     p.Surname,
		Patronymic:  p.Patronymic,
		Age:         p.Age,
		Gender:      p.Gender,
		Nationality: p.Nationality,
	}
}
