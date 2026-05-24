package mcp

import (
	"encoding/json"
	"fmt"
)

const jsonRPCVersion = "2.0"

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type toolDefinition struct {
	Name        string                 `json:"name"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type toolResult struct {
	Content           []toolContent `json:"content"`
	StructuredContent interface{}   `json:"structuredContent,omitempty"`
	IsError           bool          `json:"isError,omitempty"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func successResponse(id json.RawMessage, result interface{}) jsonRPCResponse {
	return jsonRPCResponse{JSONRPC: jsonRPCVersion, ID: id, Result: result}
}

func errorResponse(id json.RawMessage, code int, message string, data interface{}) jsonRPCResponse {
	return jsonRPCResponse{JSONRPC: jsonRPCVersion, ID: id, Error: &jsonRPCError{Code: code, Message: message, Data: data}}
}

func toolSuccess(data interface{}) toolResult {
	text, err := json.Marshal(data)
	if err != nil {
		text = []byte(fmt.Sprint(data))
	}
	return toolResult{
		Content:           []toolContent{{Type: "text", Text: string(text)}},
		StructuredContent: data,
	}
}

func toolFailure(message string, data interface{}) toolResult {
	payload := map[string]interface{}{"error": message}
	if data != nil {
		payload["details"] = data
	}
	text, _ := json.Marshal(payload)
	return toolResult{
		Content:           []toolContent{{Type: "text", Text: string(text)}},
		StructuredContent: payload,
		IsError:           true,
	}
}

func decodeParams[T any](raw json.RawMessage) (T, error) {
	var value T
	if len(raw) == 0 || string(raw) == "null" {
		return value, nil
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return value, err
	}
	return value, nil
}
