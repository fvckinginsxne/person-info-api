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

type PeopleFilters struct {
	Name        string `form:"name,omitempty" example:"John"`
	Surname     string `form:"surname,omitempty" example:"Snow"`
	Patronymic  string `form:"patronymic,omitempty" example:"Dmitrich"`
	Age         int    `form:"age,omitempty" binding:"numeric" validate:"omitempty,min=1,max=100" example:"30"`
	Gender      string `form:"gender,omitempty" validate:"omitempty,oneof=male female" example:"male"`
	Nationality string `form:"nationality,omitempty" example:"RU"`
}

type Pagination struct {
	Page int `form:"page" binding:"numeric" validate:"omitempty,min=1" example:"1"`
	Size int `form:"size" binding:"numeric" validate:"omitempty,min=1" example:"10"`
}

type SortOptions struct {
	By    string `form:"sort_by" validate:"omitempty,oneof=name surname age" example:"name"`
	Order string `form:"order,omitempty" validate:"omitempty,oneof=asc desc" example:"desc"`
}

func UpdateReqToPersonModel(p *UpdatePersonRequest) *model.Person {
	return &model.Person{
		Name:        p.Name,
		Surname:     p.Surname,
		Patronymic:  p.Patronymic,
		Age:         p.Age,
		Gender:      p.Gender,
		Nationality: p.Nationality,
	}
}

func CreateReqToPersonModel(p *CreatePersonRequest) *model.Person {
	return &model.Person{
		Name:       p.Name,
		Surname:    p.Surname,
		Patronymic: p.Patronymic,
	}
}

func ToPeopleFiltersModel(p *PeopleFilters) *model.PeopleFilters {
	return &model.PeopleFilters{
		Name:        p.Name,
		Surname:     p.Surname,
		Patronymic:  p.Patronymic,
		Age:         p.Age,
		Gender:      p.Gender,
		Nationality: p.Nationality,
	}
}

func ToPaginationModel(p *Pagination) *model.Pagination {
	return &model.Pagination{
		Page: p.Page,
		Size: p.Size,
	}
}

func ToSortOptionsModel(p *SortOptions) *model.SortOptions {
	return &model.SortOptions{
		By:    p.By,
		Order: p.Order,
	}
}
