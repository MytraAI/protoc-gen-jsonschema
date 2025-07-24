package testdata

const UnignoredMessage = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/UnignoredMessage",
    "$defs": {
        "UnignoredMessage": {
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
            "title": "Unignored Message"
        }
    }
}`
