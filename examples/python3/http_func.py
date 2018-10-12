import requests

def handle(ctx, payload):
    resp = requests.get("http://example.com")
    return {
        "status": resp.status_code,
        "headers": {k: v for (k, v) in resp.headers.items()},
        "content": resp.text
    }
