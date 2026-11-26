package httphandling

var RewriteProxyRequest = rewriteProxyRequest

var (
	UpdateAdminWithLatest = updateAdminWithLatest
	HandleAdminWebSocket  = handleWebSocket
	HandleWSMock          = handleWebSocketWSMock
)

func HasAdminConnection() bool {
	connMutex.Lock()
	defer connMutex.Unlock()

	return currentWSConnection != nil
}
