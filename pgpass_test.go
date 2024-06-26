package connection

import "testing"

func TestMatchHosts(t *testing.T) {
	tests := []struct {
		query string
		host  string
		match bool
	}{
		{"db.hiper.dk", "db.hiper.dk", true},
		{"db", "db.hiper.dk", true},
		{"dx", "db.hiper.dk", false},
		{"", "db.hiper.dk", false},
		{"hiper.dk", "db.hiper.dk", false},
	}

	for _, testCase := range tests {
		test := testCase
		t.Run(test.query, func(t *testing.T) {
			t.Parallel()
			result := hostEquals(test.query, test.host)
			if result && !test.match {
				t.Fatalf("host query %s equals %s but should not", test.query, test.host)
			}

			if !result && test.match {
				t.Fatalf("host query %s dot not equal %s but should", test.query, test.host)
			}
		})
	}
}
