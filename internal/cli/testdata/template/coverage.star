files = {
	"id": problem["id"],
	"type": problem["type"],
}
for i, sample in enumerate(problem["samples"]):
	files["samples/" + str(i) + "/input"] = sample["input"]
	files["samples/" + str(i) + "/output"] = sample["output"]
