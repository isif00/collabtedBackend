package types

type Subscription struct {
	Email                 string `json:"email"`
	Name                  string `json:"name"`
	SubscriptionPlan      string `json:"subscriptionPlan"`
	SubscriptionStatus    string `json:"subscriptionStatus"`
	SubscriptionStartDate string `json:"subscriptionStartDate"`
	SubscriptionEndDate   string `json:"subscriptionEndDate"`
	LastPaymentDate       string `json:"lastPaymentDate"`
	LastPaymentMethod     string `json:"lastPaymentMethod"`
	NextBillingDate       string `json:"nextBillingDate"`
	BillingCycle          string `json:"billingCycle"`
}
