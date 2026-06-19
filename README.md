# pucora-jsonschema
A JSON schema validator for the Pucora API Gateway

## Usage
Include in your `pucora.json` the JSON Schema configuration associated to every `endpoint` needing it. For instance:

```
{
	"version": "2",
	"endpoints": [
		{
			"endpoint": "/foo",
			"extra_config": {
				"github.com/pucora/pucora-jsonschema": {
					YOUR SCHEMA HERE
				}
			}
		}  
	]
}
```
The configuration key `"github.com/pucora/pucora-jsonschema"` takes directly as value the schema definition. 
Examples of schema can be found [here](http://json-schema.org/learn/miscellaneous-examples.html)
