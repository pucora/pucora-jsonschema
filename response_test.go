package jsonschema

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/logging"
	"github.com/pucora/lura/v2/proxy"
)

func TestResponseProxyFactory_bypass(t *testing.T) {
	// No ResponseNamespace config — should pass through without validating
	called := false
	pf := ResponseProxyFactory(logging.NoOp, proxy.FactoryFunc(func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		return func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			called = true
			return &proxy.Response{Data: map[string]interface{}{"name": "test"}}, nil
		}, nil
	}))

	p, err := pf.New(&config.EndpointConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	resp, err := p(context.Background(), &proxy.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if !called {
		t.Error("expected next proxy to be called")
	}
	if resp == nil {
		t.Error("expected a non-nil response")
	}
}

func TestResponseProxyFactory_validPass(t *testing.T) {
	// Schema requires "name" to be a string — response has name as string → should pass
	pf := ResponseProxyFactory(logging.NoOp, proxy.FactoryFunc(func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		return func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			return &proxy.Response{Data: map[string]interface{}{"name": "test"}}, nil
		}, nil
	}))

	schemaCfg := map[string]interface{}{}
	if err := json.Unmarshal([]byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		}
	}`), &schemaCfg); err != nil {
		t.Fatal(err)
	}

	p, err := pf.New(&config.EndpointConfig{
		ExtraConfig: config.ExtraConfig{
			ResponseNamespace: schemaCfg,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	resp, err := p(context.Background(), &proxy.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if resp == nil {
		t.Error("expected a non-nil response")
	}
}

func TestResponseProxyFactory_missingRequired(t *testing.T) {
	// Schema requires "age" (required field) — response does NOT have "age" → should return 500 error
	pf := ResponseProxyFactory(logging.NoOp, proxy.FactoryFunc(func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		return func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			return &proxy.Response{Data: map[string]interface{}{"name": "test"}}, nil
		}, nil
	}))

	schemaCfg := map[string]interface{}{}
	if err := json.Unmarshal([]byte(`{
		"type": "object",
		"required": ["age"],
		"properties": {
			"name": {"type": "string"},
			"age":  {"type": "integer"}
		}
	}`), &schemaCfg); err != nil {
		t.Fatal(err)
	}

	p, err := pf.New(&config.EndpointConfig{
		ExtraConfig: config.ExtraConfig{
			ResponseNamespace: schemaCfg,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	_, err = p(context.Background(), &proxy.Request{})
	if err == nil {
		t.Fatal("expected an error due to missing required field 'age'")
	}

	type statusCoder interface {
		StatusCode() int
	}
	scErr, ok := err.(statusCoder)
	if !ok {
		t.Fatalf("expected error to implement StatusCode(), got: %T", err)
	}
	if scErr.StatusCode() != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", scErr.StatusCode())
	}
}

func TestResponseProxyFactory_nilResponse(t *testing.T) {
	// When next returns nil response, the proxy should return nil without error
	pf := ResponseProxyFactory(logging.NoOp, proxy.FactoryFunc(func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		return func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			return nil, nil
		}, nil
	}))

	schemaCfg := map[string]interface{}{}
	if err := json.Unmarshal([]byte(`{"type": "object"}`), &schemaCfg); err != nil {
		t.Fatal(err)
	}

	p, err := pf.New(&config.EndpointConfig{
		ExtraConfig: config.ExtraConfig{
			ResponseNamespace: schemaCfg,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	resp, err := p(context.Background(), &proxy.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if resp != nil {
		t.Error("expected nil response to pass through")
	}
}
