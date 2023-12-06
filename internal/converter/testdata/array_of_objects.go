package testdata

const ArrayOfObjectsFail = `{
    "description": "something",
    "payload": [
        {
            "topology": "cruft"
        }
    ]
}`

const ArrayOfObjectsPass = `{
    "description": "something",
    "payload": [
        {
            "topology": "ARRAY_OF_OBJECT"
        }
    ]
}`
