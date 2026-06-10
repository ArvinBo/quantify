from __future__ import annotations

from pathlib import Path

import yaml

ROOT_DIR = Path(__file__).resolve().parent.parent.parent
CONFIG_PATH = ROOT_DIR / "config" / "default.yaml"


def load_config(path: str | Path | None = None) -> dict:
    if path is None:
        path = CONFIG_PATH
    else:
        path = Path(path)

    with open(path) as f:
        cfg = yaml.safe_load(f)

    db_path = cfg.get("db", {}).get("path", "./data/quantify.db")
    if not Path(db_path).is_absolute():
        cfg["db"]["path"] = str(ROOT_DIR / db_path)

    return cfg
