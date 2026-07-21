samples = {}
for i, sample in enumerate(problem["samples"]):
    samples[i] = {
        "input": sample["input"],
        "output": sample["output"],
    }

workspace = {
    "id": problem["id"],
    "type": problem["type"],
    "samples": samples,
}
