package domain_test

import (
	"testing"

	"ordersapi/internal/domain"
)

func TestOrderStatus_CanBeCancelled(t *testing.T) {
	tests := []struct {
		status domain.OrderStatus
		want   bool
	}{
		{domain.OrderStatusPending, true},
		{domain.OrderStatusConfirmed, false},
		{domain.OrderStatusCancelled, false},
	}
	for _, tc := range tests {
		if got := tc.status.CanBeCancelled(); got != tc.want {
			t.Errorf("%s.CanBeCancelled() = %v, se esperaba %v", tc.status, got, tc.want)
		}
	}
}
