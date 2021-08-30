package services

import (
	"com.t-systems-mms.cwa/core/security"
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/repositories"
	"context"
)

type Operators interface {
	// GetCurrentOperator gets the currently authenticated operator
	// If there is an authenticated context, but no operator exists for this subject, it will be created
	GetCurrentOperator(ctx context.Context) (domain.Operator, error)
}

type operatorsService struct {
	operators repositories.Operators
}

func NewOperatorsService(operators repositories.Operators) Operators {
	return &operatorsService{
		operators: operators,
	}
}

func (o *operatorsService) GetCurrentOperator(ctx context.Context) (domain.Operator, error) {
	token, err := security.GetTokenFromContext(ctx)
	if err != nil {
		return domain.Operator{}, err
	}

	// TODO store operator in context
	return o.operators.GetOrCreateByToken(ctx, token)
}
