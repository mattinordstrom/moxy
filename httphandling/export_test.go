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

var RedactURL = redactURL

func ResetEventLog(cleaner interface{ Cleanup(f func()) }) {
	reset := func() {
		eventLogMu.Lock()
		defer eventLogMu.Unlock()

		if eventLogFile != nil {
			eventLogFile.Close()
		}

		eventLog, eventLogSeq, eventLogFile, eventLogPath, eventLogWritten = nil, 0, nil, "", 0
	}

	reset()
	cleaner.Cleanup(reset)
}

var ForwardReq = forwardReq
