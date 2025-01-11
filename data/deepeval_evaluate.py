import os
import json
import sys
from typing import Dict, List
from deepeval.metrics import GEval, AnswerRelevancyMetric
from deepeval.test_case import LLMTestCaseParams, LLMTestCase
from dotenv import load_dotenv
from langchain_openai import AzureChatOpenAI
from client import AzureOpenAI
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns


def load_review(filepath: str) -> Dict:
    with open(filepath, 'r') as f:
        return json.load(f)

def load_results(example_id: str) -> List[Dict]:
    filepath = f'data/results/{example_id}.json'
    if os.path.exists(filepath):
        with open(filepath, 'r') as f:
            return json.load(f)
    return []

def write_result(filepath: str, review: Dict):
    with open(filepath, 'w') as f:
        json.dump(review, f, indent=4)

def evaluate_issue_recognition(expected_output: Dict, actual_output: Dict, example_id: str) -> List[Dict]:
    def get_principle_from_example_id(example_id: str) -> str:
        example_to_principle = {
            "1": "Single Responsibility Principle",
            "2": "Open/Closed Principle",
            "3": "Liskov Substitution Principle",
            "4": "Interface Segregation Principle",
            "5": "Dependency Inversion Principle",
            "6": "Don't Repeat Yourself"
        }
        return example_to_principle.get(example_id.split('_')[1], "")

    expected_principle = get_principle_from_example_id(example_id)

    actual_violations = actual_output.get('principle_violations', [])
    
    is_expected_principle_recognized = any(
        violation['principle'] == expected_principle for violation in actual_violations
    )
    
    return [{
        "example_id": example_id,
        "test_case": "Issue Recognition",
        "metric": "Recognition",
        "score": 1.0 if is_expected_principle_recognized else 0.0,
        "reason": f"{expected_principle} recognized" if is_expected_principle_recognized else f"{expected_principle} not recognized"
    }]

def evaluate_issue_description(test_cases: List[LLMTestCase], example_id: str, client: AzureOpenAI, expected_output, actual_output) -> List[Dict]:
    results = []
    metrics = [
        GEval(
            name="Correctness",
            evaluation_steps=[
                "Identify the field being compared in the test case, e.g., 'reason' or 'suggestion'.",
                "Extract the content of this field from both the actual and expected JSON outputs.",
                "Compare the content of the designated field for thematic and contextual alignment, ensuring the explanations and suggestions are addressing the same issues.",
                "Evaluate how well the actual content matches the expected content in terms of detail, clarity, and relevance to the principle violation or suggestion.",
                "Compute a similarity score from 0 to 1 based on the degree of alignment between the two field contents, with 1 indicating complete agreement and 0 indicating no agreement."
            ],
            evaluation_params=[LLMTestCaseParams.ACTUAL_OUTPUT, LLMTestCaseParams.EXPECTED_OUTPUT],
            strict_mode=False,
            model=client,
        ),
        AnswerRelevancyMetric(
           threshold=0.7,
            model=client,
            include_reason=True
        )
    ]

    for expected_violation, actual_violation, test_case in zip(expected_output['principle_violations'], actual_output['principle_violations'], test_cases):
        if (expected_violation['principle'] == actual_violation['principle']) or (test_case.name == "Principle-Test-Case"):
            for metric in metrics:
                metric_name = metric.name if hasattr(metric, 'name') else "Answer Relevancy"
                metric.measure(test_case)
                results.append({
                    "example_id": example_id,
                    "test_case": test_case.input,
                    "metric": metric_name,
                    "score": metric.score,
                    "reason": metric.reason
                })
    
    return results

def evaluate_example(example_id: str, client: AzureOpenAI) -> List[Dict]:
    expected_filepath = f'data/expected/{example_id}.json'
    actual_filepath = f'data/responses/{example_id}.json'
    result_filepath = f'data/results/{example_id}.json'

    expected_output = load_review(expected_filepath)
    actual_output = load_review(actual_filepath)

    results = []

    issue_recognition_results = evaluate_issue_recognition(expected_output, actual_output, example_id)
    results.extend(issue_recognition_results)

    test_cases = []
    min_length = min(len(expected_output['principle_violations']), len(actual_output['principle_violations']))
    
    for index in range(min_length):
        expected_violation = expected_output['principle_violations'][index]
        actual_violation = actual_output['principle_violations'][index]

        reason_test_case = LLMTestCase(
            name="Reason-Test-Case",
            input=f"Review reason for {actual_violation['principle']} in {example_id}",
            actual_output=actual_violation.get('reason', ''),
            expected_output=expected_violation.get('reason', '')
        )
        suggestion_test_case = LLMTestCase(
            name="Suggestion-Test-Case",
            input=f"Review suggestion for {actual_violation['principle']} in {example_id}",
            actual_output=actual_violation['suggestion'],
            expected_output=expected_violation['suggestion']
        )

        methods_test_case = LLMTestCase(
            name="Method-Name-Test-Case",
            input=f"Review method names for {actual_violation['principle']} in {example_id}",
            actual_output=actual_violation['method_names'],
            expected_output=expected_violation['method_names']
        )

        principle_test_case = LLMTestCase(
            name="Principle-Test-Case",
            input=f"Review principle names for {actual_violation['principle']} in {example_id}",
            actual_output=actual_violation['principle'],
            expected_output=expected_violation['principle']
        )

        test_cases.append(reason_test_case)
        test_cases.append(suggestion_test_case)
        test_cases.append(methods_test_case)
        test_cases.append(principle_test_case)

    issue_description_results = evaluate_issue_description(test_cases, example_id, client, expected_output, actual_output)
    results.extend(issue_description_results)

    write_result(result_filepath, results)
    return results

def evaluate_all_examples(client: AzureOpenAI) -> List[Dict]:
    example_ids = [f"example_{i}_{j}" for i in range(1, 7) for j in range(1, 3)]
    all_results = []

    for example_id in example_ids:
        print(f"Evaluating example: {example_id}")
        results = evaluate_example(example_id, client)
        all_results.extend(results)

    write_result('all_evaluation_results.json', all_results)
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
    sns.barplot(x="example_id", y="score", hue="metric", data=df, width=0.4)
    plt.title('Evaluation Scores by Example and Metric')
    plt.xlabel('Example ID')
    plt.ylabel('Score')
    plt.legend(title="Metric")
    plt.xticks(rotation=45)
    plt.show()

def main():
    load_dotenv()

    API_BASE = os.getenv("AZURE_OPENAI_API_BASE")
    API_KEY = os.getenv("AZURE_OPENAI_API_KEY")
    API_VERSION = os.getenv("AZURE_OPENAI_API_VERSION")
    DEPLOYMENT_NAME = os.getenv("AZURE_OPENAI_DEPLOYMENT_NAME")
    # Replace these with real values
    custom_model = AzureChatOpenAI(
        openai_api_version=API_VERSION,
        azure_deployment=DEPLOYMENT_NAME,
        azure_endpoint=API_BASE,
        openai_api_key=API_KEY,
    )
    azure_openai = AzureOpenAI(model=custom_model)
    #result = evaluate_example("example_1_1", azure_openai)
    results = load_results("example_1_1")
    visualize_results(results)

if __name__ == "__main__":
    main()