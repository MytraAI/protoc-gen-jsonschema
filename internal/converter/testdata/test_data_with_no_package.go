package testdata

const TypeNamesWithNoPackageFail = `{
    "description": "something",
    "payload": [
        {"topology": "cruft"}
    ]
}`

const TypeNamesWithNoPackagePass = `{
    "description": "something",
    "payload": [
        {"topology": "ARRAY_OF_MESSAGE"}
    ]
}`
