
from collections import defaultdict
from typing import Dict, List

import numpy as np
from common_config import COUNTER, RESULTS_DATA_DIR
from data_loader import load_json_files


def aggregate_and_calculate_stats(example_id: str, all_categories: List[str], is_java: bool, model: str) -> Dict[str, Dict[str, float]]:
    response_filepaths = [f'{RESULTS_DATA_DIR}/{model}/{example_id}_{i}_java.json' for i in range(COUNTER)] if is_java else [f'{RESULTS_DATA_DIR}/{model}/{example_id}_{i}.json' for i in range(COUNTER)]
    all_results = load_json_files(response_filepaths)

    f1_scores = []
    correctness_scores = defaultdict(list)

    for results in all_results:
        for result in results:
            test_case_scores = defaultdict(list)
            for i in range(len(result)):
                if result[i]['metric'] == 'F1-Score':
                    f1_scores.append(result[i]['score'])
                
                if result[i]['metric'] == 'Correctness':
                    test_case = result[i]['test_case']
                    
                    if "reason" in test_case.lower():
                        category = "reason"
                    elif "suggestion" in test_case.lower():
                        category = "suggestion"
                    elif "method names" in test_case.lower():
                        category = "method names"
                    elif "code segments" in test_case.lower():
                        category = "code segments"
                    else:
                        category = "other" 
                    
                    test_case_scores[category].append(result[i]['score'])
            
            for category, scores in test_case_scores.items():
                if scores: 
                    mean = np.mean(scores)
                    std = np.std(scores)
                    correctness_scores[category].append({
                        'mean': mean,
                        'std': std,
                        'scores': scores
                    })

    f1_mean = np.mean(f1_scores) if f1_scores else 0
    f1_std = np.std(f1_scores) if f1_scores else 0

    correctness_stats = {}
    for category, test_case_stats in correctness_scores.items():
        category_means = [tc['mean'] for tc in test_case_stats]
        overall_mean = np.mean(category_means) if category_means else 0
        overall_std = np.std(category_means) if category_means else 0
        correctness_stats[category] = {
            "mean": overall_mean,
            "std": overall_std,
            "test_cases": test_case_stats
        }

    for category in all_categories:
        if category not in correctness_stats:
            correctness_stats[category] = {
                "mean": 0,
                "std": 0,
                "test_cases": []
            }

    return {
        'f1': {'mean': f1_mean, 'std': f1_std, 'scores': f1_scores},
        'correctness': correctness_stats
    }