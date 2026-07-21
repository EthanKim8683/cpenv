def expand(entry):
    if type(entry) == "string":
        return {"File": {"Content": entry}}
    elif type(entry) == "dict":
        return {"Dir": {"Entries": {str(k): expand(v) for k, v in entry.items()}}}
    else:
        fail("expand: want string or dict, got " + str(type(entry)))

def encode(entry):
    if type(entry) != "dict":
        fail("encode: want dict, got " + str(type(entry)))
    return json.encode(expand(entry))
