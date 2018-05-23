# Dispatch Web Example

This is a very simple example which demonstrates building a data driven web
site backed by Dispatch functions via the API gateway.

## Prerequisites

* A Dispatch instance

## Create the Dispatch Entities

Assuming that you are starting from a bare Dispatch installation, the following
steps are required.  You may omit some of these steps depending on what has
already been configured:

1. Create the python base image:

```bash
dispatch create base-image python3-base dispatchframework/python3-base:0.0.6 --language python3
```

2. Create the python image (optionally, add any dependencies here):

```bash
dispatch create image python3 python3-base
```

3. Create the hello-py function (optionally, add schema):

    a. Create the `hello.py` file with the editor of your choice (`examples/python3/hello.py`):

    ```python
    def handle(ctx, payload):
        name = "Noone"
        place = "Nowhere"
        if payload:
            name = payload.get("name", name)
            place = payload.get("place", place)
        return {
            "myField": "Hello, %s from %s" % (name, place)
        }
    ```
    b. Register the function with Dispatch:

    ```bash
    dispatch create function --image=python3 hello-py ./examples/python3 --handler=hello.handle
    ```

4. Create an API endpoint:

```bash
dispatch create api hello hello-py --method POST --path /hello --cors
```

## Get the Host and Port

The following assumes a single host for both Dispatch and the API Gateway (this
does not hold for all installations):

```bash
export API_HOST=$(cat ~/.dispatch/config.json | jq -r .host)
export API_PORT=$(cat ~/.dispatch/config.json | jq -r .api-http-port)
```

Test the endpoints with curl (or any http client):

```bash
curl -X POST http://$API_HOST:$API_PORT/hello -H "Content-Type: application/json" -d '{"name": "Jon", "place": "Winterfell"}'
```

## Create a Web Page

Lastly we create a web page which communicates with the API endpoint we just created.

1. Create the `config.js` file. This configures the web page to point to the API gateway:

```bash
cat << EOF > config.js
var env = {
    "dispatchAPI": "http://$API_HOST:$API_PORT"
}
EOF
```

2. Create the `index.html` file with the editor of your choice (`examples/web/index.html`):

```html
<html>
<head>
    <title>Dispatch Web Example</title>
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.3.1/jquery.min.js"></script>
    <script src="config.js"></script>
</head>
<body>
    <div id="message"></div>
    <form method="POST" name="hello" id="hello">
        <p><label for="name">Name:</label>
        <input type="text" name="name" id="name"></p>

        <p><label for="place">Place:</label>
        <input type="text" name="place" id="place"></p>

        <input value="Submit" type="submit">
    </form>
    <script>
$("#hello").on("submit", function(e) {
    e.preventDefault();
    var formData = JSON.stringify({
        name: $("#hello :input[name='name']").val(),
        place: $("#hello :input[name='place']").val()
    });
    $.ajax({
        type: "POST",
        url: env.dispatchAPI + "/hello",
        data: formData,
        success: function(response){
            $("#message").html("<strong>" + response.myField + "</strong>");
        },
        contentType : "application/json"
    });
})
    </script>
</body>
</html>
```

## Open index.html with a Web Browser

You should see a very simple page with a form with fields for name and place. By
pressing the submit button, javascript is making an AJAX request to the API
gateway with the name and place values JSON encoded. Upon a successful request
the response message is returned and displayed above the form.

## Extend this Example

This was a very simple and brief introduction into building data backed web pages
with Dispatch.
