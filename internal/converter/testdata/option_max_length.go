package testdata

const OptionMaxLengthFail = `{
    "query": "abcdefghijklmnopqrstuvwxyz",
	"page_number": 4
}`

const OptionMaxLengthPass = `{
	"query": "abc",
	"page_number": 4
}`
