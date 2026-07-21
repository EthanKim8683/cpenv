def dir_from(entries):
    return {
        "dir": {
            "entries": {str(k): v for k, v in entries.items()},
        },
    }

def file_from(content):
    return {
        "file": {
            "content": str(content),
        },
    }
