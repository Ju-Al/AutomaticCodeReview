import json
import os
from typing import Dict, List

def load_json_file(filepath: str) -> Dict:
    if os.path.exists(filepath):
        with open(filepath, 'r') as f:
            return json.load(f)
    else:
        print("File doesn't exist")
        return []
    
def load_json_files(filepaths: List[str]) -> List[Dict]:
    reviews = []
    for filepath in filepaths:
        if os.path.exists(filepath):
            reviews.append(load_json_file(filepath))
    return reviews

def write_result(filepath: str, review: Dict):
    with open(filepath, 'w') as f:
        json.dump(review, f, indent=4)