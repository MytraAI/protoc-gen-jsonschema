package testdata

const MessageKind10 = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/MessageKind10",
    "$defs": {
        "MessageKind10": {
            "properties": {
                "name": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "rating": {
                    "type": "number"
                },
                "complete": {
                    "type": "boolean"
                }
            },
            "additionalProperties": true,
            "type": "object",
            "title": "Message Kind 10"
        }
    }
}`
