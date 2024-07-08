# =========================================
# Cloud Datastore
# =========================================

from datetime import datetime, timedelta
import os
import uuid

from ._common import session

LOCALRUNNER_ADDR = os.environ["LOCALRUNNER_ADDR"]


def _validate(ent: dict):
    for _, value in ent.items():
        if (isinstance(value, str) and len(value.encode()) > 1500) or (
            isinstance(value, bytes) and len(value) > 1500
        ):
            raise ValueError("Value greater than 1500 bytes", value)


def _list_all(namespace, kind):
    namespace = namespace or "$default"
    r = session.get(LOCALRUNNER_ADDR + f"/_internal/datastore/{namespace}/{kind}")
    r.raise_for_status()
    return r.json()


def _get_single(namespace, kind, key):
    namespace = namespace or "$default"
    r = session.get(LOCALRUNNER_ADDR + f"/_internal/datastore/{namespace}/{kind}/{key}")
    r.raise_for_status()
    return r.json()


def _put_single(namespace, kind, key, ent):
    namespace = namespace or "$default"
    _validate(ent)
    r = session.put(
        LOCALRUNNER_ADDR + f"/_internal/datastore/{namespace}/{kind}/{key}",
        json=ent,
    )
    r.raise_for_status()
    return r.json()


def _delete_single(namespace, kind, key):
    namespace = namespace or "$default"
    r = session.delete(
        LOCALRUNNER_ADDR + f"/_internal/datastore/{namespace}/{kind}/{key}"
    )
    r.raise_for_status()


class DataStoreEntity(dict):
    def __init__(self, key):
        super().__init__()
        self.key = key
        self.id = key.id
        self.kind = key.kind

    @staticmethod
    def create(kind: str, ent: dict):
        x = DataStoreEntity(DataStoreKey(kind, ent["#key"]))
        x.update(ent)
        return x


class DataStoreQuery:
    def __init__(self, namespace, kind):
        if not isinstance(namespace, str):
            raise TypeError(type(namespace))
        if not isinstance(kind, str):
            raise TypeError(type(kind))
        self.namespace = namespace
        self.kind = kind
        self.filters = []
        #! ORDER WORKS ONLY FOR NUMERIC FIELDS
        self.order = []

    def fetch(self, limit=None, offset=None):
        kind = _list_all(self.namespace, self.kind)

        result = []
        for ent in kind:
            ok = True
            for f, op, v in self.filters:
                if op == "=" and ent.get(f, None) != v:
                    ok = False
                    break
            if ok:
                result.append(DataStoreEntity.create(self.kind, ent))

        if self.order:
            result.sort(key=self._make_orderer())

        if offset is not None:
            result = result[offset:]

        if limit is not None:
            result = result[:limit]

        return result
    
    def cmp(self, a, b):
        return (a > b) - (a < b)

    def _make_orderer(self):
        def order_func(item):
            sort_keys = []
            for order_key in self.order:
                desc = order_key.startswith('-')
                key = order_key[1:] if desc else order_key
                value = item.get(key, 0)  # Assuming numerical values, defaulting to 0
                if desc:
                    value = -value  # Invert value for descending order
                sort_keys.append(value)
            return tuple(sort_keys)
        return order_func

    def add_filter(self, field, op, value):
        if op not in ("=",):
            raise ValueError("unknown op: " + op)
        self.filters.append((field, op, value))


class DataStoreClient:
    def __init__(self, namespace=None):
        self.namespace = namespace

    def query(self, **kwargs):
        kwargs.setdefault("namespace", self.namespace)
        return DataStoreQuery(**kwargs)

    def key(self, *parts):
        return DataStoreKey(*parts)

    def get(self, key):
        return DataStoreEntity.create(
            key.kind, _get_single(self.namespace, key.kind, key.id)
        )

    def put(self, entity):
        return _put_single(self.namespace, entity.kind, entity.id, entity)

    def delete(self, entity):
        return _delete_single(self.namespace, entity.kind, entity.id)

    def transaction(self):
        return DataStoreTransaction()


class DataStoreTransaction:
    def __enter__(self):
        return self

    def __exit__(self, *args):
        pass


class DataStoreKey:
    def __init__(self, *parts):
        if len(parts) % 2 == 1:
            # Generates a random number, by creating a random UUID
            # (e.g. `5f755fc1-08bf-4f9d-8748-1a0dfec75416`), converts
            # it to a string, takes the last 12 characters
            # (`1a0dfec75416`) and interprets that as a hex number.

            # It's like this because I needed a random string first,
            # but later needed a random number. And some parts of the
            # code use strings/integers interchangeably, so I am not
            # sure.
            parts = parts + (int(str(uuid.uuid4())[-12:], base=16),)
        self.parts = parts
        self.kind = parts[-2]
        self.id = parts[-1]

    def to_legacy_urlsafe(self):
        return ":".join(self.parts).encode()


def datastore_create_client(namespace=None):
    return DataStoreClient(namespace=namespace)


def datastore_parse_legacy_key(key):

    return DataStoreKey(*key.split(":"))


def datastore_create_entity(key, **_kwargs):
    return DataStoreEntity(key)
