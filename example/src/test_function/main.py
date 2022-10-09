import os
import cloud

ENV_NAME = os.environ["ENV_NAME"]


def main(message, context):
    print("hello", message)
    return "hello from " + ENV_NAME, 200


print("initialized", ENV_NAME)
