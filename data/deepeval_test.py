import json
import os
from typing import Dict, List
from dotenv import load_dotenv
import matplotlib.pyplot as plt
import seaborn as sns
import pandas as pd
from deepeval.metrics import GEval
from deepeval.test_case import LLMTestCaseParams, LLMTestCase

def load_review(filepath: str) -> Dict:
    with open(filepath, 'r') as f:
        return json.load(f)
    
def load_results(filepath: str) -> List[Dict]:
    if not os.path.exists(filepath):
        return []
    with open(filepath, 'r') as f:
        return json.load(f)

def write_result(filepath: str, review: Dict):
    with open(filepath, 'w') as f:
        json.dump(review, f, indent=4)

def evaluate_example(example_id: str):
    expected_filepath = f'data/expected/{example_id}.json'
    actual_filepath = f'data/responses/{example_id}.json'
    #result_filepath = f'data/results/{example_id}.json'

    expected_output = load_review(expected_filepath)
    actual_output = load_review(actual_filepath)

    results = []

    test_cases = []

    for index in range(len(expected_output['principle_violations'])):
        expected_violation = expected_output['principle_violations'][index]

        if index < len(actual_output['principle_violations']):
            actual_violation = actual_output['principle_violations'][index]
        else:
            # Wenn weniger expected_violations als actual_violations
            actual_violation = expected_violation  # Placeholder, falls weniger expected_violations

        reason_test_case = LLMTestCase(
            input=f"Review reason for {actual_violation['principle']} in {example_id}",
            actual_output=actual_violation['reason'], 
            expected_output=expected_violation['reason']
        )

        suggestion_test_case = LLMTestCase(
            input=f"Review suggestion for {actual_violation['principle']} in {example_id}",
            actual_output=actual_violation['suggestion'],
            expected_output=expected_violation['suggestion']
        )

        methods_test_case = LLMTestCase(
            input=f"Review method names for {actual_violation['principle']} in {example_id}",
            actual_output=actual_violation['method_names'],
            expected_output=expected_violation['method_names']
        )

        test_cases.append(reason_test_case)
        test_cases.append(suggestion_test_case)
        test_cases.append(methods_test_case)

    metrics = [
        GEval(
            name="Correctness",
            criteria="Correctness - determine if the actual output is correct according to the expected output.",
            evaluation_params=[LLMTestCaseParams.ACTUAL_OUTPUT, LLMTestCaseParams.EXPECTED_OUTPUT],
            strict_mode=True
        ),
        GEval(
            name="Relevance",
            criteria="Relevance - determine if the actual output is relevant to the given context.",
            evaluation_params=[LLMTestCaseParams.ACTUAL_OUTPUT, LLMTestCaseParams.EXPECTED_OUTPUT],
            strict_mode=True
        ),
        GEval(
            name="Clarity",
            criteria="Clarity - determine if the actual output is clear and easy to understand.",
            evaluation_params=[LLMTestCaseParams.ACTUAL_OUTPUT, LLMTestCaseParams.EXPECTED_OUTPUT],
            strict_mode=True
        ),
        GEval(
            name="Actionability",
            criteria="Actionability - determine if the suggestions are actionable and practical.",
            evaluation_params=[LLMTestCaseParams.ACTUAL_OUTPUT, LLMTestCaseParams.EXPECTED_OUTPUT],
            strict_mode=True
        )
    ]
    for test_case in test_cases:
            metric = metrics[0]
            metric.measure(test_case)
            results.append({
                "example_id": example_id,
                "test_case": test_case.input,
                "metric": metric.name,
                "score": metric.score,
                "reason": metric.reason
            })

    """ for metric in metrics:
        for test_case in test_cases:
            metric.measure(test_case)
            results.append({
                "example_id": example_id,
                "test_case": test_case.input,
                "metric": metric.name,
                "score": metric.score,
                "reason": metric.reason
            })
 """
    #write_result(results, result_filepath)
    return results

def evaluate_all_examples() -> List[Dict]:
    example_ids = [f"example_{i}_{j}" for i in range(1, 7) for j in range(1, 3)]
    all_results = []

    for example_id in example_ids:
        print(f"Evaluating example: {example_id}")
        results = evaluate_example(example_id)
        all_results.extend(results)

    return all_results

def visualize_results(results: List[Dict]):
    data = {
        "example_id": [],
        "metric": [],
        "score": []
    }
    
    for result in results:
        data["example_id"].append(result["example_id"])
        data["metric"].append(result["metric"])
        data["score"].append(result["score"])
    
    df = pd.DataFrame(data)
    
    plt.figure(figsize=(12, 6))
    sns.barplot(x="example_id", y="score", hue="metric", data=df)
    plt.title('Evaluation Scores by Example and Metric')
    plt.xlabel('Example ID')
    plt.ylabel('Score')
    plt.legend(title="Metric")
    plt.xticks(rotation=45)
    plt.show()

def visualize_single_example(example_id: str, results: List[Dict]):
    example_results = [result for result in results if result["example_id"] == example_id]
    metrics = set(result["metric"] for result in example_results)
    
    for metric in metrics:
        data = {
            "test_case": [],
            "score": []
        }
        
        for result in example_results:
            if result["metric"] == metric:
                data["test_case"].append(result["test_case"])
                data["score"].append(result["score"])
        
        df = pd.DataFrame(data)
        
        plt.figure(figsize=(12, 6))
        sns.barplot(x="test_case", y="score", data=df)
        plt.title(f'Evaluation Scores for {metric} in {example_id}')
        plt.xlabel('Test Case')
        plt.ylabel('Score')
        plt.xticks(rotation=45)
        plt.ylim(0, 1)
        plt.show()

def main():
    load_dotenv()

    API_KEY = os.getenv("AZURE_OPENAI_API_KEY")

    results_filepath = 'evaluation_results.json'
    
    if os.path.exists(results_filepath):
        all_results = load_results(results_filepath)
    else:
        all_results = evaluate_all_examples()
        write_result(all_results, results_filepath)

    #visualize_single_example("example_1_1", all_results)
    """ results_filepath = 'all_evaluation_results.json'
    
    if os.path.exists(results_filepath):
        all_results = load_results(results_filepath)
    else:
        all_results = evaluate_all_examples(client)

    example_ids = [f"example_{i}_{j}" for i in range(1, 7) for j in range(1, 3)]
    for example_id in example_ids:
        visualize_single_example(example_id, all_results) """

if __name__ == "__main__":
    main()