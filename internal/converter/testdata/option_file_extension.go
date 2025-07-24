package testdata

const OptionFileExtension = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/OptionFileExtension",
    "$defs": {
        "OptionFileExtension": {
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
            "title": "Option File Extension"
        }
    }
}`
