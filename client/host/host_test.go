package host

import (
	"context"
	"testing"
)

func TestFromContext(t *testing.T) {
	tc := []struct {
		name    string
		cf      func() context.Context
		expHost string
		expOk   bool
	}{
		{
			name: "valid",
			cf: func() context.Context {
				return NewContext(context.TODO(), "1.2.3.4")
			},
			expHost: "1.2.3.4",
			expOk:   true,
		},
		{
			name: "blank",
			cf: func() context.Context {
				return NewContext(context.TODO(), "")
			},
			expHost: "",
			expOk:   true,
		},
		{
			name: "wrong type",
			cf: func() context.Context {
				return context.WithValue(context.TODO(), hostKey, 123)
			},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			ctx := c.cf()

			h, ok := FromContext(ctx)
			if c.expOk != ok {
				t.Fatalf("context value ok: expected %v got %v", c.expOk, ok)
			}
			if c.expHost != h {
				t.Fatalf("context value: expected %q got %q", c.expHost, h)
			}
		})
	}
}
