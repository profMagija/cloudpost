"""
A module which wraps all the GCP functionality so it
can easily be mocked/simulated
"""

from .datastore import (
    datastore_create_client,
    datastore_parse_legacy_key,
    datastore_create_entity,
)
from .pubsub import pubsub_create_publisher
from .storage import storage_create_client
