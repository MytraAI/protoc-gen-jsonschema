package testdata

const Proto2NestedObjectFail = `{
	"payload": {
		"topology": "FLAT"	
	}
}`

const Proto2NestedObjectPass = `{
	"description": "lots of attributes",
	"payload": {
		"name": "something",
		"timestamp": "1970-01-01T00:00:00Z",
		"id": 1,
		"rating": 100,
		"complete": true,
		"topology": "FLAT"
	}
}`
