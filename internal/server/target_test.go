package server

import (
	"testing"
)

// https://docs.aws.amazon.com/ja_jp/AmazonRDS/latest/AuroraUserGuide/Aurora.Overview.Endpoints.html#Aurora.Overview.Endpoints.Types
func TestGetRdsEndpointType(t *testing.T) {
	type testCase struct {
		input    string
		expected rdsEndpointType
	}

	testCases := []testCase{
		{
			input:    "mydbcluster.cluster-123456789012.us-east-1.rds.amazonaws.com:3306",
			expected: cluster,
		},
		{
			input:    "mydbcluster.cluster-ro-123456789012.us-east-1.rds.amazonaws.com:3306",
			expected: cluster,
		},
		{
			input:    "mydbinstance.cluster-ro-123456789012.us-east-1.rds.amazonaws.com:3306",
			expected: cluster,
		},
		{
			input:    "mydbinstance.123456789012.us-east-1.rds.amazonaws.com:3306",
			expected: instance,
		},
		{
			input:    "mydbcluster.123456789012.us-east-1.rds.amazonaws.com:3306",
			expected: instance,
		},
	}

	for _, tc := range testCases {
		actual := getRdsEndpointType(tc.input)
		if actual != tc.expected {
			t.Errorf("for input %s, expected %v but got %v", tc.input, tc.expected, actual)
		}
	}
}
