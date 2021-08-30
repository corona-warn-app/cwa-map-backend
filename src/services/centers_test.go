package services

import (
	"com.t-systems-mms.cwa/core/security"
	"com.t-systems-mms.cwa/domain"
	mocks2 "com.t-systems-mms.cwa/mocks/external/geocoding"
	mocks "com.t-systems-mms.cwa/mocks/repositories"
	"com.t-systems-mms.cwa/mocks/services"
	"context"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"testing"
)

func TestCentersService_ImportCenters(t *testing.T) {
	service, centersRepository, operatorService, operatorsRepository, _ := createCentersService()

	subject := "TestCentersService_ImportCenters"
	ctx := authenticatedContext(subject)
	ref := "TestCentersService_ImportCenters"

	centers := []domain.Center{
		{
			Name:          "Testcenter",
			UserReference: &ref,
			Website:       nil,
			Operator:      nil,
			OperatorUUID:  "",
			Address:       "",
			AddressNote:   nil,
			OpeningHours:  nil,
			Appointment:   nil,
			TestKinds:     nil,
			EnterDate:     nil,
			LeaveDate:     nil,
			DCC:           nil,
		},
	}

	operatorsRepository.On("GetOrCreateByToken", ctx, subject).Return(domain.Operator{UUID: "operatorUUID"}, nil)
	operatorService.On("GetCurrentOperator", ctx).Return(domain.Operator{UUID: "operatorUUID"}, nil)

	centersRepository.On("SaveMultiple", ctx, centers).Run(func(args mock.Arguments) {
		centers := args[1].([]domain.Center)
		for i, _ := range centers {
			centers[i].UUID = "generatedUUID"
		}
	}).Return(centers, nil)

	centersRepository.On("FindByOperatorAndUserReference", ctx, "operatorUUID", "TestCentersService_ImportCenters").
		Return(domain.Center{}, gorm.ErrRecordNotFound)

	result, err := service.ImportCenters(ctx, centers, true)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "generatedUUID", result[0].UUID)
}

func TestCentersService_UpdateCenter(t *testing.T) {
	service, centersRepository, operatorsRepository, _, _ := createCentersService()
	subject := "TestCentersService_UpdateCenter"
	ctx := authenticatedContext(subject)
	ref := "TestCentersService_UpdateCenter"
	centers := []domain.Center{
		{
			Name:          "Testcenter",
			UserReference: &ref,
			Website:       nil,
			Operator:      nil,
			OperatorUUID:  "",
			Address:       "",
			AddressNote:   nil,
			OpeningHours:  nil,
			Appointment:   nil,
			TestKinds:     nil,
			EnterDate:     nil,
			LeaveDate:     nil,
			DCC:           nil,
		},
	}

	operatorsRepository.On("GetOrCreateByToken", ctx, subject).
		Return(domain.Operator{UUID: "operatorUUID"}, nil)

	centersRepository.On("SaveMultiple", ctx, centers).
		Return(centers, nil)

	centersRepository.On("FindByOperatorAndUserReference", ctx, "operatorUUID", subject).
		Return(domain.Center{UUID: "existingUUID"}, nil)

	result, err := service.ImportCenters(ctx, centers, true)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, result) {
		return
	}
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "existingUUID", result[0].UUID)
}

func TestCentersService_ImportUnauthorized(t *testing.T) {
	service, _, _, _, _ := createCentersService()
	result, err := service.ImportCenters(context.Background(), nil, true)
	assert.ErrorIs(t, err, security.ErrUnauthorized)
	assert.Nil(t, result)
}

func authenticatedContext(subject string) context.Context {
	token := jwt.New()
	_ = token.Set(jwt.SubjectKey, subject)
	return context.WithValue(context.Background(), jwtauth.TokenCtxKey, token)
}

func createCentersService() (Centers, *mocks.Centers, *mocks.Operators, *services.Operators, *mocks2.Geocoder) {
	centersRepository := &mocks.Centers{}
	operatorsRepository := &mocks.Operators{}
	operatorService := &services.Operators{}
	geocoder := &mocks2.Geocoder{}
	service := NewCentersService(centersRepository, operatorsRepository, operatorService, geocoder)
	return service, centersRepository, operatorsRepository, operatorService, geocoder
}
