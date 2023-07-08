import json
import os
import cloud

ENV_NAME = os.environ["ENV_NAME"]


def main(message, context):
    print("hello", message)
    return "hello from " + ENV_NAME, 200


print(
    json.dumps(
        {
            "message": "initialized",
            "fields": {
                "ENV_NAME": ENV_NAME,
            },
        }
    )
)
