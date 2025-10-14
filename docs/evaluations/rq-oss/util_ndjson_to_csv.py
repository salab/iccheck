import json
import csv
import sys

if len(sys.argv) < 3:
    raise "usage: script.py ndjson_file csv_file"

# File paths
ndjson_file = sys.argv[1]
csv_file = sys.argv[2]

# Open the NDJSON file and the CSV file
with open(ndjson_file, 'r', encoding='utf-8') as ndjson_f, open(csv_file, 'w', newline='', encoding='utf-8') as csv_f:
    # Initialize the CSV writer
    writer = csv.DictWriter(csv_f, fieldnames=None)

    # Read and parse each line of the NDJSON file
    first_line = True
    for line in ndjson_f:
        # Load the JSON object from the line
        data = json.loads(line.strip())

        # Write headers (fieldnames) on the first line
        if first_line:
            writer.fieldnames = data.keys()
            writer.writeheader()  # Write header only once
            first_line = False

        # Write the data row
        writer.writerow(data)
