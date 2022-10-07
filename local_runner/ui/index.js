
function $(x) {
    return document.getElementById(x)
}

let curThing = {
    function: null,
    container: null,
    queue: null,
};

function init_thing(things, name) {
    $(name + "-status").innerText = ""

    things.forEach(f => {
        let li = document.createElement("li");
        li.innerText = f
        if (f == curThing[name]) {
            li.style.fontWeight = "bold"
        }
        li.onclick = () => {
            curThing[name] = f;
            init_thing(things, name);
        }
        $(name + "-status").appendChild(li);
    })
}

async function send_function_request(endpoint) {
    if (curThing.function == null) {
        alert("select function first");
        return
    }

    $("function-response").innerText = "Running ..."

    let payload = $("function-payload").value;
    const resp = await fetch(`/${curThing.function}/${endpoint}`, {
        body: payload,
        method: "post",
    })
    $("function-response").innerText = await resp.text()
}

function init_storage(storage) {
    let status = $("storage-status");
    status.innerHTML = "";
    Object.keys(storage).forEach(bucket => {
        let bucket_li = document.createElement("li");
        bucket_li.appendChild(document.createTextNode(bucket))
        let bucket_ul = bucket_li.appendChild(document.createElement("ul"));
        let ar = storage[bucket];
        ar.sort();
        ar.forEach(object => {
            let obj_li = bucket_ul.appendChild(document.createElement("li"));
            let obj_a = obj_li.appendChild(document.createElement("a"));
            obj_a.innerText = "/" + object;
            obj_a.href = `/_internal/storage/${bucket}/${object}`
            obj_a.target = "_blank"
        })
        status.appendChild(bucket_li);
    })
}

function init_datastore(ds) {
    let status = $("datastore-status");
    status.innerHTML = "";
    let view = $("datastore-view");
    Object.keys(ds).forEach(namespace => {
        let ns_li = document.createElement("li");
        ns_li.appendChild(document.createTextNode(namespace))
        let ns_ul = ns_li.appendChild(document.createElement("ul"));
        Object.keys(ds[namespace]).forEach(kind => {
            let kd_li = ns_ul.appendChild(document.createElement("li"));
            kd_li.appendChild(document.createTextNode(kind))
            let kd_ul = kd_li.appendChild(document.createElement("ul"));
            Object.keys(ds[namespace][kind]).forEach(key => {
                let key_li = kd_ul.appendChild(document.createElement("li"));
                key_li.innerText = key;
                key_li.onclick = () => {
                    view.innerText = JSON.stringify(ds[namespace][kind][key], null, 2)
                }
            })
        })
        status.appendChild(ns_li);
    })
}

async function send_container_request() {
    if (!curThing.container) {
        alert("select container first");
        return
    }

    $("container-response").innerText = "Running ..."

    let payload = $("container-payload").value;
    let endpoint = $("container-path").value;
    const resp = await fetch(`/${curThing.container}/${endpoint}`, {
        body: payload,
        method: $("container-method").value,
    })
    $("container-response").innerText = await resp.text()
}
async function send_queue_request() {
    if (!curThing.queue) {
        alert("select queue first");
        return
    }

    $("queue-response").innerText = "Running ..."

    let payload = $("queue-payload").value;
    const resp = await fetch(`/_internal/queue/${curThing.queue}/publish`, {
        body: payload,
        method: "post",
    })
    $("queue-response").innerText = await resp.text()
}

async function load_data() {
    try {
        $("server-status").innerText = "Loading...";
        const resp = await fetch("/_internal/info");
        const data = await resp.json();
        $("server-status").innerText = "Online";

        init_thing(data.Functions, 'function');
        init_thing(data.Containers, 'container');
        init_thing(Object.keys(data.Queues), 'queue');
        init_datastore(data.Datastore)
        init_storage(data.Storage)

        setTimeout(load_data, 3000);
    } catch (e) {
        console.error(e);
        $("server-status").innerText = "Error (" + e.toString() + ")";
    }
}

load_data()