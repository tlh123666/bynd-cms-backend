package httpx

import (
	"net/http"
	"testing"
)

func TestUpstreamStatusDoesNotInvalidateCMSSession(t *testing.T) {
	tests := []struct {
		upstream int
		want     int
	}{
		{http.StatusUnauthorized, http.StatusBadGateway},
		{http.StatusForbidden, http.StatusBadGateway},
		{http.StatusNotFound, http.StatusNotFound},
		{http.StatusConflict, http.StatusConflict},
	}

	for _, test := range tests {
		if got := UpstreamStatus(test.upstream); got != test.want {
			t.Fatalf("UpstreamStatus(%d) = %d, want %d", test.upstream, got, test.want)
		}
	}
}
