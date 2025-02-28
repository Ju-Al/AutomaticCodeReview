from langchain_community.callbacks import get_openai_callback
from typing import Dict, List
from deepeval.metrics import GEval
from deepeval.test_case import LLMTestCase
from client import AzureOpenAI
from common_config import *
from data_loader import load_json_file, load_json_files, write_result
from test_cases import generate_test_cases
from metrics import get_all_metrics, select_metrics
from typing import List, Dict


def evaluate_test_cases_with_metrics(test_cases: List[LLMTestCase], metrics: List[GEval], example_id: str) -> List[Dict]:
    results = []
    total_tokens = 0

    for test_case in test_cases:
        for metric in metrics:
            with get_openai_callback() as cb:
                metric.measure(test_case)
                tokens = cb.total_tokens
                total_tokens += tokens
            
            results.append({
                "example_id": example_id,
                "test_case": test_case.input,
                "test_case_category": test_case.name,
                "metric": metric.name,
                "score": metric.score,
                "reason": metric.reason,
                "tokens_used": tokens
            })
    return results

def evaluate_example(example_id: str, client: AzureOpenAI, is_java: bool) -> List[Dict]:
    expected_filepath = f'{EXPECTED_DATA_DIR}/{example_id}_java.json' if is_java else f'{EXPECTED_DATA_DIR}/{example_id}.json'
    response_filepaths = [f'{RESPONSES_DATA_DIR}/{example_id}_{i}_java.json' for i in range(COUNTER)] if is_java else [f'{RESPONSES_DATA_DIR}/{example_id}_{i}.json' for i in range(COUNTER)]

    expected_output = load_json_file(expected_filepath)
    actual_outputs = load_json_files(response_filepaths)

    all_expected_principles = [violation['principle'] for violation in expected_output['principle_violations']]

    all_results = []

    for i, actual_output in enumerate(actual_outputs):
        all_actual_principles = [violation['principle'] for violation in actual_output['principle_violations']]
        test_cases_f1 = [LLMTestCase(
                        name=f"F1-Test-Case-{example_id}_{i}",
                        input="Check principle violations in the document.",
                        actual_output=all_actual_principles,
                        expected_output=all_expected_principles
                    )]

        all_metrics = get_all_metrics(client)
        selected_metrics = select_metrics(all_metrics, ["F1-Score"])

        results = []
        results.append(evaluate_test_cases_with_metrics(test_cases_f1, selected_metrics, f'{example_id}_{i}'))

        test_cases = generate_test_cases(expected_output, actual_output, example_id)
        selected_metrics = select_metrics(all_metrics, ["Correctness"])
        results.append(evaluate_test_cases_with_metrics(test_cases, selected_metrics,  f'{example_id}_{i}'))

        result_filepath = f'{RESULTS_DATA_DIR}/{example_id}_{i}_java.json' if is_java else f'{RESULTS_DATA_DIR}/{example_id}_{i}.json'
        write_result(result_filepath, results)
        all_results.append(results)
    return all_results 