package testdata

const SharedMessage = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/SharedMessage",
    "$defs": {
        "SharedMessage": {
            "properties": {
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "version": {
                    "type": "integer"
                }
            },
            "additionalProperties": true,
            "type": "object",
            "title": "Shared Message"
        }
    }
}`

const ConsumerMessage = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/ConsumerMessage",
    "$defs": {
        "ConsumerMessage": {
            "properties": {
                "consumer_id": {
                    "type": "string"
                },
                "shared_data": {
                    "$ref": "#/$defs/shared.SharedMessage",
                    "additionalProperties": true
                },
                "status": {
                    "oneOf": [
                        {
                            "type": "string"
                        },
                        {
                            "type": "integer"
                        }
                    ],
                    "enum": [
                        "UNKNOWN",
                        0,
                        "ACTIVE",
                        1,
                        "INACTIVE",
                        2
                    ]
                },
                "shared_list": {
                    "items": {
                        "$ref": "#/$defs/shared.SharedMessage"
                    },
                    "type": "array"
                }
            },
            "additionalProperties": true,
            "type": "object",
            "title": "Consumer Message"
        },
        "shared.SharedMessage": {
            "properties": {
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "version": {
                    "type": "integer"
                }
            },
            "additionalProperties": true,
            "type": "object",
            "title": "Shared Message"
        }
    }
}`

const ConsumerMessageWithDepdendenciesAsRefs = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/ConsumerMessage",
    "$defs": {
        "ConsumerMessage": {
            "properties": {
                "consumer_id": {
                    "type": "string"
                },
                "shared_data": {
                    "$ref": "SharedMessage.json#/$defs/SharedMessage",
                    "additionalProperties": true
                },
                "status": {
                    "oneOf": [
                        {
                            "type": "string"
                        },
                        {
                            "type": "integer"
                        }
                    ],
                    "enum": [
                        "UNKNOWN",
                        0,
                        "ACTIVE",
                        1,
                        "INACTIVE",
                        2
                    ]
                },
                "shared_list": {
                    "items": {
                        "$ref": "SharedMessage.json#/$defs/SharedMessage"
                    },
                    "type": "array"
                }
            },
            "additionalProperties": true,
            "type": "object",
            "title": "Consumer Message"
        }
    }
}`

const SharedMessageWithDepdendenciesAsRefs = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/$defs/SharedMessage",
    "$defs": {
        "SharedMessage": {
            "properties": {
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "version": {
                    "type": "integer"
                }
            },
            "additionalProperties": true,
            "type": "object",
            "title": "Shared Message"
        }
    }
}`

const ConsumerMessagePass = `{
    "consumer_id": "test123",
    "shared_data": {
        "id": "shared1",
        "name": "Test Shared",
        "version": 1
    },
    "status": "ACTIVE",
    "shared_list": [
        {
            "id": "shared2",
            "name": "Another Shared",
            "version": 2
        }
    ]
}`

const ConsumerMessageFail = `{
    "consumer_id": 123,
    "shared_data": {
        "id": "shared1",
        "name": "Test Shared",
        "version": "not_a_number"
    },
    "status": "INVALID_STATUS"
}`

const SharedMessagePass = `{
    "id": "test123",
    "name": "Test Message",
    "version": 1
}`

const SharedMessageFail = `{
    "id": 123,
    "name": "Test Message",
    "version": "not_a_number"
}`
