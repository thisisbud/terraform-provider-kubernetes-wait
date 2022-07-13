package provider

import "testing"

func Test_doIt(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := doIt(); (err != nil) != tt.wantErr {
				t.Errorf("doIt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
