package models

import (
	"testing"
)

func TestCreateSubscriptionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateSubscriptionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateSubscriptionRequest{
				ServiceName: "Test Service",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			wantErr: false,
		},
		{
			name: "valid request with end date",
			req: CreateSubscriptionRequest{
				ServiceName: "Test Service",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
				EndDate:     strPtr("12-2024"),
			},
			wantErr: false,
		},
		{
			name: "empty service name",
			req: CreateSubscriptionRequest{
				ServiceName: "",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			wantErr: true,
		},
		{
			name: "zero price",
			req: CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       0,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			wantErr: true,
		},
		{
			name: "negative price",
			req: CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       -100,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			wantErr: true,
		},
		{
			name: "empty user id",
			req: CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       299,
				UserID:      "",
				StartDate:   "01-2024",
			},
			wantErr: true,
		},
		{
			name: "invalid user id",
			req: CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       299,
				UserID:      "not-a-uuid",
				StartDate:   "01-2024",
			},
			wantErr: true,
		},
		{
			name: "empty start date",
			req: CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "",
			},
			wantErr: true,
		},
		{
			name: "invalid start date format",
			req: CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "2024-01",
			},
			wantErr: true,
		},
		{
			name: "end date before start date",
			req: CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "06-2024",
				EndDate:     strPtr("01-2024"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestUpdateSubscriptionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateSubscriptionRequest
		wantErr bool
	}{
		{
			name:    "empty request - valid (partial update)",
			req:     UpdateSubscriptionRequest{},
			wantErr: false,
		},
		{
			name: "valid price update",
			req: UpdateSubscriptionRequest{
				Price: intPtr(399),
			},
			wantErr: false,
		},
		{
			name: "valid service name update",
			req: UpdateSubscriptionRequest{
				ServiceName: strPtr("New Name"),
			},
			wantErr: false,
		},
		{
			name: "empty service name",
			req: UpdateSubscriptionRequest{
				ServiceName: strPtr(""),
			},
			wantErr: true,
		},
		{
			name: "whitespace service name",
			req: UpdateSubscriptionRequest{
				ServiceName: strPtr("   "),
			},
			wantErr: true,
		},
		{
			name: "zero price",
			req: UpdateSubscriptionRequest{
				Price: intPtr(0),
			},
			wantErr: true,
		},
		{
			name: "negative price",
			req: UpdateSubscriptionRequest{
				Price: intPtr(-100),
			},
			wantErr: true,
		},
		{
			name: "invalid start date format",
			req: UpdateSubscriptionRequest{
				StartDate: strPtr("2024-01"),
			},
			wantErr: true,
		},
		{
			name: "valid dates update",
			req: UpdateSubscriptionRequest{
				StartDate: strPtr("01-2024"),
				EndDate:   strPtr("12-2024"),
			},
			wantErr: false,
		},
		{
			name: "clear end date with empty string",
			req: UpdateSubscriptionRequest{
				EndDate: strPtr(""),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestCreateSubscriptionRequest_ParseDates(t *testing.T) {
	tests := []struct {
		name       string
		req        CreateSubscriptionRequest
		wantErr    bool
		wantEndNil bool
	}{
		{
			name: "start date only",
			req: CreateSubscriptionRequest{
				StartDate: "01-2024",
			},
			wantErr:    false,
			wantEndNil: true,
		},
		{
			name: "both dates",
			req: CreateSubscriptionRequest{
				StartDate: "01-2024",
				EndDate:   strPtr("12-2024"),
			},
			wantErr:    false,
			wantEndNil: false,
		},
		{
			name: "end before start",
			req: CreateSubscriptionRequest{
				StartDate: "06-2024",
				EndDate:   strPtr("01-2024"),
			},
			wantErr: true,
		},
		{
			name: "invalid start date",
			req: CreateSubscriptionRequest{
				StartDate: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := tt.req.ParseDates()

			if tt.wantErr {
				if err == nil {
					t.Error("ParseDates() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDates() unexpected error: %v", err)
				return
			}

			if start.IsZero() {
				t.Error("ParseDates() start date is zero")
			}

			if tt.wantEndNil && end != nil {
				t.Error("ParseDates() expected end to be nil")
			}
			if !tt.wantEndNil && end == nil {
				t.Error("ParseDates() expected end to be non-nil")
			}
		})
	}
}

func TestUpdateSubscriptionRequest_ParseDates(t *testing.T) {
	tests := []struct {
		name         string
		req          UpdateSubscriptionRequest
		wantErr      bool
		wantStartNil bool
		wantEndNil   bool
		wantClearEnd bool
	}{
		{
			name:         "empty request",
			req:          UpdateSubscriptionRequest{},
			wantErr:      false,
			wantStartNil: true,
			wantEndNil:   true,
			wantClearEnd: false,
		},
		{
			name: "only start date",
			req: UpdateSubscriptionRequest{
				StartDate: strPtr("01-2024"),
			},
			wantErr:      false,
			wantStartNil: false,
			wantEndNil:   true,
			wantClearEnd: false,
		},
		{
			name: "both dates",
			req: UpdateSubscriptionRequest{
				StartDate: strPtr("01-2024"),
				EndDate:   strPtr("12-2024"),
			},
			wantErr:      false,
			wantStartNil: false,
			wantEndNil:   false,
			wantClearEnd: false,
		},
		{
			name: "clear end date",
			req: UpdateSubscriptionRequest{
				EndDate: strPtr(""),
			},
			wantErr:      false,
			wantStartNil: true,
			wantEndNil:   true,
			wantClearEnd: true,
		},
		{
			name: "invalid start date",
			req: UpdateSubscriptionRequest{
				StartDate: strPtr("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, clearEnd, err := tt.req.ParseDates()

			if tt.wantErr {
				if err == nil {
					t.Error("ParseDates() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDates() unexpected error: %v", err)
				return
			}

			if tt.wantStartNil && start != nil {
				t.Error("ParseDates() expected start to be nil")
			}
			if !tt.wantStartNil && start == nil {
				t.Error("ParseDates() expected start to be non-nil")
			}

			if tt.wantEndNil && end != nil {
				t.Error("ParseDates() expected end to be nil")
			}
			if !tt.wantEndNil && end == nil {
				t.Error("ParseDates() expected end to be non-nil")
			}

			if clearEnd != tt.wantClearEnd {
				t.Errorf("ParseDates() clearEnd = %v, want %v", clearEnd, tt.wantClearEnd)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
