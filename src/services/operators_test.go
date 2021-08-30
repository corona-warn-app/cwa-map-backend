package services

import (
	"com.t-systems-mms.cwa/core/security"
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/mocks/repositories"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOperatorsService_GetCurrentOperator(t *testing.T) {
	service, repo := createOperatorsService()
	subject := "TestOperatorsService_GetCurrentOperator"
	ctx := authenticatedContext(subject)

	repo.Mock.On("GetOrCreateByToken", ctx, subject).Return(
		domain.Operator{UUID: "TestOperatorsService_GetCurrentOperator"}, nil)

	operator, err := service.GetCurrentOperator(ctx)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "TestOperatorsService_GetCurrentOperator", operator.UUID)
}

func TestOperatorsService_GetCurrentOperatorUnauthenticated(t *testing.T) {
	service, _ := createOperatorsService()
	ctx := context.Background()

	_, err := service.GetCurrentOperator(ctx)
	assert.ErrorIs(t, err, security.ErrUnauthorized)
}

func createOperatorsService() (Operators, *repositories.Operators) {
	operatorsRepository := &repositories.Operators{}
	operatorsService := NewOperatorsService(operatorsRepository)
	return operatorsService, operatorsRepository
}
