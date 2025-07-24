package testdata

const SecondMessage = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/SecondMessage",
    "$defs": {
        "SecondMessage": {
            "properties": {
                "name2": {
                    "type": "string"
                },
                "timestamp2": {
                    "type": "string"
                },
                "id2": {
                    "type": "integer"
                },
                "rating2": {
                    "type": "number"
                },
                "complete2": {
                    "type": "boolean"
                }
            },
            "additionalProperties": true,
            "type": "object",
            "title": "Second Message"
        }
    }
}`

const SecondMessageFail = `{"complete2": "hello"}`

const SecondMessagePass = `{"complete2": true}`
