package caddyapi

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func Test_GetServer_ConfigPathMissing(t *testing.T) {
	c := NewClient("http://localhost", nil)
	httpmock.ActivateNonDefault(c.client.GetClient())
	defer httpmock.DeactivateAndReset()

	mockRequest("GET", "http://localhost/config/apps/http/servers/https", "", `{"error":"invalid traversal path at: config/apps/http"}`, http.StatusBadRequest)

	_, err := c.GetServer("@config/apps/http/servers/https")
	assert.ErrorIs(t, err, ErrConfigPathNotFound)
}
