package services

import "github.com/CollabTED/CollabTed-Backend/pkg/types"

type SubscriptionService struct {
	AuthService *AuthService
}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{
		AuthService: NewAuthService(),
	}
}

func (s *SubscriptionService) GetUserSubscription(email string) (any, error) {
	userModel, err := s.AuthService.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	subscription := types.Subscription{
		Email:              userModel.Email,
		Name:               userModel.Name,
		SubscriptionPlan:   string(userModel.SubscriptionPlan),
		SubscriptionStatus: string(userModel.SubscriptionStatus),
		BillingCycle:       string(userModel.BillingCycle),
	}

	return subscription, nil
}
