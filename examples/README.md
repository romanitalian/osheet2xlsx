# Examples

This folder contains minimal sample inputs and outputs for osheet2xlsx.

- `sample.osheet`: a tiny ZIP-based osheet with a single text file entry (demonstration of fallback modes)
- `typed.osheet`: an osheet with `document.json` using typed cells, including a formula and a date
- `outputs/`: generated `.xlsx` files after running conversion examples below

## How to run

```bash
# Convert all .osheet files from this folder into ./examples/outputs
osheet2xlsx convert ./examples --pattern "*.osheet" --out-dir ./examples/outputs --overwrite --progress

# Inspect a specific osheet
osheet2xlsx inspect ./examples/typed.osheet --json

# Validate structure
osheet2xlsx validate ./examples/typed.osheet --json
```

## Create typed.osheet

We provide a prebuilt `typed.osheet`, but you can recreate it with:

```bash
mkdir -p /tmp/os2x && cat > /tmp/os2x/document.json <<'JSON'
{
  "sheets": [
    {
      "name": "S",
      "cells": [
        [{"t":"n","v":1},{"t":"n","v":2}],
        [{"f":"SUM(A1:B1)"}, "2024-01-02"]
      ]
    }
  ]
}
JSON
(cd /tmp/os2x && zip -rq ./typed.osheet .) && mv /tmp/os2x/typed.osheet ./examples/
```
