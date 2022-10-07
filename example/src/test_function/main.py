import cloud


def main(message, context):
    print("started")
    client = cloud.datastore_create_client("test")
    key = client.key("TestKind", 100)
    entity = cloud.datastore_create_entity(key)
    entity["x"] = 1
    entity["y"] = "z"
    print("putting")
    client.put(entity)
    print("put")
    print("getting")
    ent2 = client.get(key)
    print("got", ent2)

    items = list(client.query(kind="TestKind").fetch())
    print(items)

    return "ok", 200


print("initialized!")
