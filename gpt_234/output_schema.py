output_schema = {
    "type": "object",
    "properties": {
        "type": {
            "type": "string"
        },
        "principle_violations": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "principle": {
                        "type": "string"
                    },
                    "lines": {
                        "type": "array",
                        "items": {"type": "integer"}
                    },
                    "reason": {"type": "string"},
                    "suggestion": {"type": "string"}
                },
                "required": ["principle", "lines", "reason", "suggestion"],
                "additionalProperties": False
            }
        },
        "overall_feedback": {"type": "string"},
        "refusal": {"type": ["string", "null"]}
    },
    "required": ["type", "principle_violations", "overall_feedback", "refusal"],
    "additionalProperties": False
}