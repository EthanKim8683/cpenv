files = {
    "id": problem["id"],
    "type": problem["type"],
}
for i, sample in enumerate(problem["samples"]):
    files["samples/%d/input" % i] = sample["input"]
    files["samples/%d/output" % i] = sample["output"]
