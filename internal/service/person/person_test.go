package person

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"person-info/internal/domain/model"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SavePerson(ctx context.Context, person *model.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *MockStorage) Exists(ctx context.Context, person *model.Person) (bool, error) {
	args := m.Called(ctx, person)
	return args.Bool(0), args.Error(1)
}

type MockAgeProvider struct {
	mock.Mock
}

func (m *MockAgeProvider) Age(ctx context.Context, name string) (int, error) {
	args := m.Called(ctx, name)
	return args.Int(0), args.Error(1)
}

type MockGenderProvider struct {
	mock.Mock
}

func (m *MockGenderProvider) Gender(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}

type MockNationalityProvider struct {
	mock.Mock
}

func (m *MockNationalityProvider) Nationality(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}

func TestService_Save(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		inputPerson    *model.Person
		mockSetup      func(*MockStorage, *MockAgeProvider, *MockGenderProvider, *MockNationalityProvider)
		expectedResult *model.Person
		expectedError  error
	}{
		{
			name: "successful save",
			inputPerson: &model.Person{
				Name:    "John",
				Surname: "Snow",
			},
			mockSetup: func(ms *MockStorage, ma *MockAgeProvider, mg *MockGenderProvider, mn *MockNationalityProvider) {
				ms.On("Exists", mock.Anything, &model.Person{Name: "John", Surname: "Snow"}).Return(false, nil)
				ma.On("Age", mock.Anything, "John").Return(30, nil)
				mg.On("Gender", mock.Anything, "John").Return("male", nil)
				mn.On("Nationality", mock.Anything, "John").Return("US", nil)
				ms.On("SavePerson", mock.Anything, &model.Person{
					Name:        "John",
					Surname:     "Snow",
					Age:         30,
					Gender:      "male",
					Nationality: "US",
				}).Return(nil)
			},
			expectedResult: &model.Person{
				Name:        "John",
				Surname:     "Snow",
				Age:         30,
				Gender:      "male",
				Nationality: "US",
			},
			expectedError: nil,
		},
		{
			name: "person already exists",
			inputPerson: &model.Person{
				Name:    "John",
				Surname: "Snow",
			},
			mockSetup: func(ms *MockStorage, ma *MockAgeProvider, mg *MockGenderProvider, mn *MockNationalityProvider) {
				ms.On("Exists", mock.Anything, &model.Person{Name: "John", Surname: "Snow"}).Return(true, nil)
			},
			expectedResult: nil,
			expectedError:  fmt.Errorf("service.person.Save: %w", ErrPersonExists),
		},
		{
			name: "storage exists check error",
			inputPerson: &model.Person{
				Name:    "John",
				Surname: "Snow",
			},
			mockSetup: func(ms *MockStorage, ma *MockAgeProvider, mg *MockGenderProvider, mn *MockNationalityProvider) {
				ms.On("Exists", mock.Anything, &model.Person{Name: "John", Surname: "Snow"}).Return(false, errors.New("db error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("service.person.Save: db error"),
		},
		{
			name: "agify provider error",
			inputPerson: &model.Person{
				Name:    "John",
				Surname: "Snow",
			},
			mockSetup: func(ms *MockStorage, ma *MockAgeProvider, mg *MockGenderProvider, mn *MockNationalityProvider) {
				ms.On("Exists", mock.Anything, &model.Person{Name: "John", Surname: "Snow"}).Return(false, nil)
				ma.On("Age", mock.Anything, "John").Return(0, errors.New("agify api error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("service.person.Save: agify api error"),
		},
		{
			name: "genderize provider error",
			inputPerson: &model.Person{
				Name:    "John",
				Surname: "Snow",
			},
			mockSetup: func(ms *MockStorage, ma *MockAgeProvider, mg *MockGenderProvider, mn *MockNationalityProvider) {
				ms.On("Exists", mock.Anything, &model.Person{Name: "John", Surname: "Snow"}).Return(false, nil)
				ma.On("Age", mock.Anything, "John").Return(30, nil)
				mg.On("Gender", mock.Anything, "John").Return("", errors.New("genderize api error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("service.person.Save: genderize api error"),
		},
		{
			name: "nationalize provider error",
			inputPerson: &model.Person{
				Name:    "John",
				Surname: "Snow",
			},
			mockSetup: func(ms *MockStorage, ma *MockAgeProvider, mg *MockGenderProvider, mn *MockNationalityProvider) {
				ms.On("Exists", mock.Anything, &model.Person{Name: "John", Surname: "Snow"}).Return(false, nil)
				ma.On("Age", mock.Anything, "John").Return(30, nil)
				mg.On("Gender", mock.Anything, "John").Return("male", nil)
				mn.On("Nationality", mock.Anything, "John").Return("", errors.New("nationalize api error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("service.person.Save: nationalize api error"),
		},
		{
			name: "save person error",
			inputPerson: &model.Person{
				Name:    "John",
				Surname: "Snow",
			},
			mockSetup: func(ms *MockStorage, ma *MockAgeProvider, mg *MockGenderProvider, mn *MockNationalityProvider) {
				ms.On("Exists", mock.Anything, &model.Person{Name: "John", Surname: "Snow"}).Return(false, nil)
				ma.On("Age", mock.Anything, "John").Return(30, nil)
				mg.On("Gender", mock.Anything, "John").Return("male", nil)
				mn.On("Nationality", mock.Anything, "John").Return("US", nil)
				ms.On("SavePerson", mock.Anything, mock.Anything).Return(errors.New("save error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("service.person.Save: save error"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStorage := new(MockStorage)
			mockAgeProvider := new(MockAgeProvider)
			mockGenderProvider := new(MockGenderProvider)
			mockNationalityProvider := new(MockNationalityProvider)

			tt.mockSetup(mockStorage, mockAgeProvider, mockGenderProvider, mockNationalityProvider)

			service := New(
				logger,
				mockStorage,
				mockAgeProvider,
				mockGenderProvider,
				mockNationalityProvider,
			)

			result, err := service.Save(context.Background(), tt.inputPerson)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, result)

			mockStorage.AssertExpectations(t)
			mockAgeProvider.AssertExpectations(t)
			mockGenderProvider.AssertExpectations(t)
			mockNationalityProvider.AssertExpectations(t)
		})
	}
}
