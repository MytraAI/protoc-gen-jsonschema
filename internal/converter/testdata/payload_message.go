package testdata

const PayloadMessage = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/PayloadMessage",
    "$defs": {
        "PayloadMessage": {
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
                },
                "topology": {
                    "enum": [
                        "FLAT",
                        0,
                        "NESTED_OBJECT",
                        1,
                        "NESTED_MESSAGE",
                        2,
                        "ARRAY_OF_TYPE",
                        3,
                        "ARRAY_OF_OBJECT",
                        4,
                        "ARRAY_OF_MESSAGE",
                        5
                    ],
                    "oneOf": [
                        {
                            "type": "string"
                        },
                        {
                            "type": "integer"
                        }
                    ],
                    "title": "Topology"
                }
            },
            "additionalProperties": true,
            "type": "object",
            "title": "Payload Message"
        }
    }
}`

const PayloadMessageFail = `{"topology": "MAP"}`

const PayloadMessagePass = `{"topology": "ARRAY_OF_MESSAGE"}`

const PayloadMessageWithDepdendenciesAsRefs = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/PayloadMessage",
    "$defs": {
        "PayloadMessage": {
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
                },
                "topology": {
                    "oneOf": [
                        {
                            "type": "string"
                        },
                        {
                            "type": "integer"
                        }
                    ],
                    "enum": [
                        "FLAT",
                        0,
                        "NESTED_OBJECT",
                        1,
                        "NESTED_MESSAGE",
                        2,
                        "ARRAY_OF_TYPE",
                        3,
                        "ARRAY_OF_OBJECT",
                        4,
                        "ARRAY_OF_MESSAGE",
                        5
                    ]
                }
            },
            "additionalProperties": true,
            "type": "object",
            "title": "Payload Message"
        }
    }
}`
