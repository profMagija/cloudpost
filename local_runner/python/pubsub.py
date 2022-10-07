# =========================================
# Cloud Pub/Sub
# =========================================

from base64 import b64encode
import os
from ._common import session

LOCALRUNNER_ADDR = os.environ["LOCALRUNNER_ADDR"]


class _Result:
    def __init__(self, res):
        self._res = res

    def result(self):
        return self._res


def _post_message(name, data):
    r = session.post(LOCALRUNNER_ADDR + f"/_internal/queue/{name}/publish", data=data)
    r.raise_for_status()


class PubsubPublisherClient:
    def publish(self, topic_name, data):
        if not isinstance(data, bytes):
            raise TypeError("topic data must be bytes")
        _post_message(topic_name, data)
        return _Result("123456789")

    def topic_path(self, project_id, topic_name):
        return topic_name


def pubsub_create_publisher():
    return PubsubPublisherClient()
