def handle(ctx, payload):
    name = payload.get("name", "Noone")  
    place = payload.get("place", "Nowhere")
    return {"myField": "Hello, %s from %s" % (name, place)}
