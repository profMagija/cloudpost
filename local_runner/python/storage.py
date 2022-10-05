# =========================================
# Cloud Storage
# =========================================

import logging
import os
import requests

LOCALRUNNER_ADDR = os.environ["LOCALRUNNER_ADDR"]


def _get_single(bucket, path):
    resp = requests.get(LOCALRUNNER_ADDR + f"/_internal/storage/{bucket}/{path}")
    resp.raise_for_status()
    return resp.content


def _get_exists(bucket, path):
    resp = requests.get(LOCALRUNNER_ADDR + f"/_internal/storage/{bucket}/{path}")
    return resp.status_code == 200


def _put_single(bucket, path, data):
    # ? Most likely windows specific problem with decoding
    try:
        requests.put(
            LOCALRUNNER_ADDR + f"/_internal/storage/{bucket}/{path}",
            data=data,
        )
    except UnicodeEncodeError:
        logging.info("Unicode encoding failed, trying latin-1...")
        data = data.encode("latin-1", "ignore").decode("latin-1", "ignore")
        requests.put(
            LOCALRUNNER_ADDR + f"/_internal/storage/{bucket}/{path}", data=data
        )


def _list_all(bucket):
    return requests.get(LOCALRUNNER_ADDR + f"/_internal/storage/{bucket}").json()


class CloudStorageClient:
    def bucket(self, name):
        return CloudStorageBucket(name)

    def get_bucket(self, name):
        return self.bucket(name)


class CloudStorageBucket:
    def __init__(self, name):
        if not isinstance(name, str):
            raise TypeError(type(name))
        self.name = name

    def blob(self, blob_name):
        return CloudStorageBlob(self, blob_name)

    def get_blob(self, blob_name):
        return CloudStorageBlob(self, blob_name)

    def list_blobs(self):
        return [CloudStorageBlob(self, name) for name in _list_all(self.name)]


class CloudStorageBlob:
    def __init__(self, bucket, name):
        if not isinstance(bucket, CloudStorageBucket):
            raise TypeError(type(bucket))
        if not isinstance(name, str):
            raise TypeError(type(name))
        self.bucket = bucket
        self.name = name

    def upload_from_string(self, data):
        if isinstance(data, str):
            data = data.encode()
        if not isinstance(data, bytes):
            raise TypeError(type(data))
        _put_single(self.bucket.name, self.name, data)

    def download_as_bytes(self):
        return _get_single(self.bucket.name, self.name)

    def download_as_text(self):
        return self.download_as_bytes().decode()

    def exists(self):
        return _get_exists(self.bucket.name, self.name)

    def __repr__(self):
        return f"<CloudStorageBlob bucket={self.bucket.name!r}, name={self.name!r}>"


def storage_create_client():
    return CloudStorageClient()
