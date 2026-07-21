samples = {}
for i, sample in enumerate(problem["samples"]):
    samples[i] = dir_from({
        "input": file_from(sample["input"]),
        "output": file_from(sample["output"]),
    })

workspace = dir_from({
    "id": file_from(problem["id"]),
    "type": file_from(problem["type"]),
    "samples": dir_from(samples),
})
