# velonetics-jsonschema
A JSON schema validator for the Velonetics API Gateway

## Usage
Include in your `velonetics.json` the JSON Schema configuration associated to every `endpoint` needing it. For instance:

```
{
	"version": "2",
	"endpoints": [
		{
			"endpoint": "/foo",
			"extra_config": {
				"github.com/velonetics/velonetics-jsonschema": {
					YOUR SCHEMA HERE
				}
			}
		}  
	]
}
```
The configuration key `"github.com/velonetics/velonetics-jsonschema"` takes directly as value the schema definition. 
Examples of schema can be found [here](http://json-schema.org/learn/miscellaneous-examples.html)
