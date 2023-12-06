package testdata

const ValidationOptionsFail = `{
	"stringWithLengthConstraints": "this string is way too long",
	"luckyNumbersWithArrayConstraints": [1]
}`

const ValidationOptionsPass = `{
	"stringWithLengthConstraints": "thisisok",
	"luckyNumbersWithArrayConstraints": [1,2,3,4]
}`
