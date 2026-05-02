package management

import "testing"

func TestRepository_GetSessionBootstrap(t *testing.T) {
	requireManagementTestEnv(t)

	t.Run("returns mapped domain payload on 200", func(t *testing.T) {
		t.Skip("TODO: assert bootstrap response maps to domain.SessionBootstrap")
	})

	t.Run("returns not found on 404", func(t *testing.T) {
		t.Skip("TODO: assert 404 maps to ErrSessionNotFound")
	})

	t.Run("returns unauthorized on 401", func(t *testing.T) {
		t.Skip("TODO: assert 401 maps to ErrUnauthorized")
	})

	t.Run("returns forbidden on 403", func(t *testing.T) {
		t.Skip("TODO: assert 403 maps to ErrForbidden")
	})

	t.Run("returns upstream unavailable on 5xx", func(t *testing.T) {
		t.Skip("TODO: assert 5xx maps to ErrUpstreamUnavailable")
	})

	t.Run("returns upstream unavailable on timeout transport error", func(t *testing.T) {
		t.Skip("TODO: assert client timeout/transport error maps to ErrUpstreamUnavailable")
	})

	t.Run("returns invalid response on malformed json", func(t *testing.T) {
		t.Skip("TODO: assert malformed payload maps to ErrInvalidResponse")
	})
}

func TestRepository_ReportSessionStatus(t *testing.T) {
	requireManagementTestEnv(t)

	t.Run("accepts 200", func(t *testing.T) {
		t.Skip("TODO: assert 200 response returns nil")
	})

	t.Run("accepts 204", func(t *testing.T) {
		t.Skip("TODO: assert 204 response returns nil")
	})

	t.Run("maps conflict to already finished", func(t *testing.T) {
		t.Skip("TODO: assert 409 or already_finished maps to ErrAlreadyFinished")
	})

	t.Run("returns not found on 404", func(t *testing.T) {
		t.Skip("TODO: assert 404 maps to ErrSessionNotFound")
	})

	t.Run("returns unauthorized on 401", func(t *testing.T) {
		t.Skip("TODO: assert 401 maps to ErrUnauthorized")
	})

	t.Run("returns forbidden on 403", func(t *testing.T) {
		t.Skip("TODO: assert 403 maps to ErrForbidden")
	})

	t.Run("returns upstream unavailable on 5xx", func(t *testing.T) {
		t.Skip("TODO: assert 5xx maps to ErrUpstreamUnavailable")
	})

	t.Run("returns upstream unavailable on timeout transport error", func(t *testing.T) {
		t.Skip("TODO: assert client timeout/transport error maps to ErrUpstreamUnavailable")
	})
}

func TestRepository_ReportSessionResults(t *testing.T) {
	requireManagementTestEnv(t)

	t.Run("accepts 200", func(t *testing.T) {
		t.Skip("TODO: assert 200 response returns nil")
	})

	t.Run("accepts 204", func(t *testing.T) {
		t.Skip("TODO: assert 204 response returns nil")
	})

	t.Run("maps conflict to already finished", func(t *testing.T) {
		t.Skip("TODO: assert 409 or already_finished maps to ErrAlreadyFinished")
	})

	t.Run("returns not found on 404", func(t *testing.T) {
		t.Skip("TODO: assert 404 maps to ErrSessionNotFound")
	})

	t.Run("returns unauthorized on 401", func(t *testing.T) {
		t.Skip("TODO: assert 401 maps to ErrUnauthorized")
	})

	t.Run("returns forbidden on 403", func(t *testing.T) {
		t.Skip("TODO: assert 403 maps to ErrForbidden")
	})

	t.Run("returns upstream unavailable on 5xx", func(t *testing.T) {
		t.Skip("TODO: assert 5xx maps to ErrUpstreamUnavailable")
	})

	t.Run("returns upstream unavailable on timeout transport error", func(t *testing.T) {
		t.Skip("TODO: assert client timeout/transport error maps to ErrUpstreamUnavailable")
	})

	t.Run("maps unexpected statuses", func(t *testing.T) {
		t.Skip("TODO: assert non-success non-mapped statuses return ErrUnexpectedStatus")
	})
}

func TestRepository_InternalHeaders(t *testing.T) {
	requireManagementTestEnv(t)

	t.Run("sends X-Internal-Service and X-Internal-Token", func(t *testing.T) {
		t.Skip("TODO: assert internal headers are present on outgoing requests")
	})

	t.Run("sends content type for patch and put", func(t *testing.T) {
		t.Skip("TODO: assert Content-Type: application/json is present for PATCH/PUT payload requests")
	})
}
