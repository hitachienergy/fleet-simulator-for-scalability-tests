import pandas as pd


def load_logs(file):
    with open(file, "r") as f:
        logs_text = [l.split(" - ") for l in f.readlines() if len(l.split(" - ")) == 4]
    logs = pd.DataFrame(logs_text, columns=["datetime", "name", "level", "message"])
    logs = logs[logs.message.apply(lambda y: y.startswith("#####"))]
    logs["message"] = logs.message.apply(
        lambda y: y.replace("#", "").strip()[:15] + "..."
    )
    logs.drop(columns=["name", "level"], inplace=True)
    logs["datetime"] = pd.to_datetime(logs["datetime"])
    return logs
