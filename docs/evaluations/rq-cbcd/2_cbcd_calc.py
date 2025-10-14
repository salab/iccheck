import json

with open("0_cbcd-dataset.json") as f:
    cbcd_dataset = json.load(f)

with open("1_stdout.log") as f:
    valid_instances = 0
    valid_precision = 0
    valid_recall = 0

    valid_pred = 0
    valid_tp = 0
    valid_oracle = 0

    for i, line in enumerate(f):
        exp = json.loads(line)
        # Since ICCheck finds "missing changes" in the same commit, cases where query and answer commit differ is not suitable for iccheck evaluation
        lookup_eq = exp["experiment"]["lookup_eq"]
        if not lookup_eq or lookup_eq == "false":
            continue
        oracle = cbcd_dataset["queries"][str(i+1)]["answers"]

        def is_correct(pred):
            for answer in oracle:
                if answer["sline"] == pred["start_l"] and answer["eline"] == pred["end_l"]:
                    return True
            return False

        predictions = exp["result"][0]["missing"] if len(exp["result"]) > 0 else []
        tp = [c for c in predictions if is_correct(c)]
        precision = len(tp) / len(predictions) if len(predictions) > 0 else 0
        recall = len(tp) / len(oracle) if len(oracle) > 0 else 0
        print(f"exp {i+1}: pred {len(predictions)}, oracle {len(oracle)}, tp {len(tp)}, precision = {precision}, recall = {recall}")

        valid_instances += 1
        valid_precision += precision
        valid_recall += recall
        valid_pred += len(predictions)
        valid_tp += len(tp)
        valid_oracle += len(oracle)

    print("==> Summary (direct average of precision / recall of each experiment instance)")
    precision = valid_precision / valid_instances
    recall = valid_recall / valid_instances
    print(f"precision = {precision}, recall = {recall}, total experiments = {valid_instances}")

    print("==> Summary (sum of all clones in all experiments)")
    precision = valid_tp / valid_pred
    recall = valid_tp / valid_oracle
    print(f"precision = {precision} ({valid_tp} / {valid_pred}), recall = {recall} ({valid_tp} / {valid_oracle})")
