package jsonschema

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/logging"
	"github.com/pucora/lura/v2/proxy"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ResponseNamespace is the key to use to store and access the response JSON Schema config
const ResponseNamespace = "validation/response-json-schema"

// ResponseProxyFactory creates a proxy factory over the injected one adding a JSON Schema
// validator middleware for RESPONSE bodies to the pipe when required.
func ResponseProxyFactory(logger logging.Logger, pf proxy.Factory) proxy.FactoryFunc {
	return proxy.FactoryFunc(func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		next, err := pf.New(cfg)
		if err != nil {
			return proxy.NoopProxy, err
		}

		jschema := responseConfigGetter(cfg.ExtraConfig)
		if jschema == nil {
			return next, nil
		}

		c := jsonschema.NewCompiler()
		c.AddResource("./response-schema.json", jschema)
		s, err := c.Compile("./response-schema.json")
		if err != nil {
			logger.Error("[ENDPOINT: " + cfg.Endpoint + "][ResponseJSONSchema] Parsing the definition:" + err.Error())
			return next, nil
		}

		logger.Debug("[ENDPOINT: " + cfg.Endpoint + "][ResponseJSONSchema] Validator enabled")
		return newResponseProxy(s, next), nil
	})
}

func newResponseProxy(schema *jsonschema.Schema, next proxy.Proxy) proxy.Proxy {
	return func(ctx context.Context, req *proxy.Request) (*proxy.Response, error) {
		resp, err := next(ctx, req)
		if err != nil {
			return resp, err
		}
		if resp == nil {
			return resp, nil
		}

		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(resp.Data); err != nil {
			return nil, &responseValidationError{err: err}
		}

		b, err := jsonschema.UnmarshalJSON(buf)
		if err != nil {
			return nil, &responseValidationError{err: err}
		}

		if err := schema.Validate(b); err != nil {
			return nil, &responseValidationError{err: err}
		}

		return resp, nil
	}
}

func responseConfigGetter(cfg config.ExtraConfig) interface{} {
	v, ok := cfg[ResponseNamespace]
	if !ok {
		return nil
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return nil
	}
	schema, err := jsonschema.UnmarshalJSON(buf)
	if err != nil {
		return nil
	}
	return schema
}

type responseValidationError struct {
	err error
}

func (r *responseValidationError) Error() string {
	return r.err.Error()
}

func (*responseValidationError) StatusCode() int {
	return 500
}
