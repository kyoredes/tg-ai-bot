package utils

import (
	"encoding/json"
	"gateway/internal/dto"
	"gateway/internal/logging"
	"io"

	"go.uber.org/zap"
	"resty.dev/v3"
)

func MakeRequest(restyClient *resty.Client, request *dto.Request, ch chan dto.Response) {
	logger := logging.Logger
	logger.Debug("making request", zap.String("url", request.URL), zap.String("method", request.Method))

	req := restyClient.R()
	if request.Body != nil {
		req.SetBody(request.Body)
	}
	if request.Headers != nil {
		req.SetHeaders(request.Headers)
	}

	res, err := req.Execute(request.Method, request.URL)
	if err != nil {
		logger.Error("error making request", zap.Error(err))
		ch <- dto.Response{Success: false}
		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Error("error reading response body")
		ch <- dto.Response{Success: false}
		return
	}

	if len(resBody) == 0 {
		logger.Error("response body is empty")
		ch <- dto.Response{Success: false}
		return
	}

	if res.StatusCode() != request.ExpectedStatusCode {
		logger.Error("status code mismatch",
			zap.Int("got", res.StatusCode()),
			zap.Int("expected", request.ExpectedStatusCode),
			zap.String("body", string(resBody)),
		)
		ch <- dto.Response{Success: false, StatusCode: res.StatusCode()}
		return
	}

	var body map[string]any
	if err := json.Unmarshal(resBody, &body); err != nil {
		logger.Error("error unmarshalling response body", zap.Error(err))
		ch <- dto.Response{Success: false}
		return
	}

	ch <- dto.Response{
		Success:    true,
		Body:       body,
		Headers:    res.Header(),
		StatusCode: res.StatusCode(),
	}
}
