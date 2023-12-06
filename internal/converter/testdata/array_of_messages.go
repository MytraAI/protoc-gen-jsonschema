package testdata

const ArrayOfMessagesFail = `{
    "description": "something",
    "payload": [
        {"topology": "cruft"}
    ]
}`

const ArrayOfMessagesPass = `{
    "description": "something",
    "payload": [
        {"topology": "ARRAY_OF_MESSAGE"}
    ]
}`
