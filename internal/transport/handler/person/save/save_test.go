package save

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	httpClient "person-info/internal/client/person"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"person-info/internal/domain/model"
	personSevice "person-info/internal/service/person"
)

type MockPersonSaver struct {
	mock.Mock
}

func (m *MockPersonSaver) Save(ctx context.Context, person *model.Person) (*model.Person, error) {
	args := m.Called(ctx, person)
	return args.Get(0).(*model.Person), args.Error(1)
}

func TestSaveHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockPersonSaver)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful save",
			requestBody: Request{
				Name:    "John",
				Surname: "Snow",
			},
			mockSetup: func(m *MockPersonSaver) {
				m.On("Save", mock.Anything, &model.Person{
					Name:    "John",
					Surname: "Snow",
				}).Return(&model.Person{
					Name:        "John",
					Surname:     "Snow",
					Age:         30,
					Gender:      "male",
					Nationality: "RU",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"name":"John","surname":"Snow","patronymic":"","agify":30,"genderize":"male","nationalize":"RU"}`,
		},
		{
			name:           "empty request body",
			requestBody:    nil,
			mockSetup:      func(m *MockPersonSaver) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"request body is empty"}`,
		},
		{
			name:           "invalid request - missing required fields",
			requestBody:    Request{Name: "John"},
			mockSetup:      func(m *MockPersonSaver) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "invalid person name",
			requestBody: Request{
				Name:    "фффф",
				Surname: "Snow",
			},
			mockSetup: func(m *MockPersonSaver) {
				m.On("Save", mock.Anything, &model.Person{
					Name:    "фффф",
					Surname: "Snow",
				}).Return((*model.Person)(nil), httpClient.ErrInvalidName)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid name"}`,
		},
		{
			name: "person already exists",
			requestBody: Request{
				Name:    "John",
				Surname: "Snow",
			},
			mockSetup: func(m *MockPersonSaver) {
				m.On("Save", mock.Anything, &model.Person{
					Name:    "John",
					Surname: "Snow",
				}).Return((*model.Person)(nil), personSevice.ErrPersonExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"person already exists"}`,
		},
		{
			name: "save error - database failure",
			requestBody: Request{
				Name:    "John",
				Surname: "Doe",
			},
			mockSetup: func(m *MockPersonSaver) {
				m.On("Save", mock.Anything, &model.Person{
					Name:    "John",
					Surname: "Doe",
				}).Return((*model.Person)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
		{
			name: "with patronymic",
			requestBody: Request{
				Name:       "John",
				Surname:    "Snow",
				Patronymic: "Ivanovich",
			},
			mockSetup: func(m *MockPersonSaver) {
				m.On("Save", mock.Anything, &model.Person{
					Name:       "John",
					Surname:    "Snow",
					Patronymic: "Ivanovich",
				}).Return(&model.Person{
					Name:        "John",
					Surname:     "Snow",
					Patronymic:  "Ivanovich",
					Age:         25,
					Gender:      "male",
					Nationality: "RU",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"name":"John","surname":"Snow","patronymic":"Ivanovich","agify":25,"genderize":"male","nationalize":"RU"}`,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSaver := new(MockPersonSaver)
			tt.mockSetup(mockSaver)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Подготавливаем запрос
			if tt.requestBody != nil {
				jsonBody, _ := json.Marshal(tt.requestBody)
				c.Request = httptest.NewRequest(
					"POST",
					"/persons",
					bytes.NewBuffer(jsonBody),
				)
				c.Request.Header.Set("Content-Type", "application/json")
			} else {
				c.Request = httptest.NewRequest(
					"POST",
					"/persons",
					nil,
				)
			}

			handler := New(context.Background(), logger, mockSaver)
			handler(c)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code mismatch")

			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String(), "response body mismatch")
			}

			mockSaver.AssertExpectations(t)
		})
	}
}
