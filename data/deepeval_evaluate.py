from collections import defaultdict
from itertools import groupby
import os
import json
from typing import Dict, List
from deepeval.metrics import GEval, AnswerRelevancyMetric
from deepeval.test_case import LLMTestCaseParams, LLMTestCase
from dotenv import load_dotenv
from langchain_openai import AzureChatOpenAI
import numpy as np
from client import AzureOpenAI
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns


def load_review(filepath: str) -> Dict:
    print(filepath)
    with open(filepath, 'r') as f:
        return json.load(f)
    
def load_reviews(filepaths: List[str]) -> List[Dict]:
    reviews = []
    for filepath in filepaths:
        if os.path.exists(filepath):
            with open(filepath, 'r') as f:
                reviews.append(json.load(f))
    return reviews

def load_results(example_id: str) -> List[Dict]:
    filepath = f'data/results/{example_id}.json'
    if os.path.exists(filepath):
        with open(filepath, 'r') as f:
            return json.load(f)
    return []

def write_result(filepath: str, review: Dict):
    with open(filepath, 'w') as f:
        json.dump(review, f, indent=4)

def get_all_metrics(client: AzureOpenAI) -> List[GEval]:
    return [
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
        GEval(
            name="F1-Score",
            criteria="Calculate the F1-Score based on the provided lists of principle violations.",
            evaluation_steps=[
                "Identify True Positives (TP) as principles that are in both the actual and expected violation lists.",
                "Identify False Positives (FP) as principles that are in the actual violation list but not in the expected violation list.",
                "Identify False Negatives (FN) as principles that are in the expected violation list but not in the actual violation list.",
                "Calculate Precision as TP / (TP + FP).",
                "Calculate Recall as TP / (TP + FN).",
                "Calculate F1-Score as 2 * (Precision * Recall) / (Precision + Recall)."
            ],
            evaluation_params=[LLMTestCaseParams.ACTUAL_OUTPUT, LLMTestCaseParams.EXPECTED_OUTPUT],
            strict_mode=False,
            model=client
        )
    ]

def select_metrics(all_metrics: List[GEval], metrics_to_select: List[str]) -> List[GEval]:
    return [metric for metric in all_metrics if metric.name in metrics_to_select]

def select_test_cases(test_cases: List[LLMTestCase], test_case_names: List[str]) -> List[LLMTestCase]:
    return [test_case for test_case in test_cases if test_case.name in test_case_names]

def generate_test_cases(expected_output: Dict, actual_output: Dict, example_id: str) -> List[LLMTestCase]:
    test_cases = []
    
    expected_violations_dict = {violation['principle']: violation for violation in expected_output['principle_violations']}
    actual_violations_dict = {violation['principle']: violation for violation in actual_output['principle_violations']}

    all_principles = set(expected_violations_dict.keys()).union(set(actual_violations_dict.keys()))
    
    for principle in all_principles:
        expected_violation = expected_violations_dict.get(principle, {})
        actual_violation = actual_violations_dict.get(principle, {})
        
        # Check if principle exists in both expected and actual
        if principle in expected_violations_dict and principle in actual_violations_dict:
            # Create test cases for reason, suggestion, method names
            reason_test_case = LLMTestCase(
                name=f"Reason-Test-Case-{principle}",
                input=f"Review reason for {principle} in {example_id}",
                actual_output=actual_violation.get('reason', ''),
                expected_output=expected_violation.get('reason', '')
            )
            suggestion_test_case = LLMTestCase(
                name=f"Suggestion-Test-Case-{principle}",
                input=f"Review suggestion for {principle} in {example_id}",
                actual_output=actual_violation.get('suggestion', ''),
                expected_output=expected_violation.get('suggestion', '')
            )
            methods_test_case = LLMTestCase(
                name=f"Method-Name-Test-Case-{principle}",
                input=f"Review method names for {principle} in {example_id}",
                actual_output=actual_violation.get('method_names', []),
                expected_output=expected_violation.get('method_names', [])
            )
            code_segment_test_case = LLMTestCase(
                name=f"Changes-Test-Case-{principle}",
                input=f"Review code segments for {principle} in {example_id}",
                actual_output=actual_violation.get('changes', []),
                expected_output=expected_violation.get('changes', [])
            )
            
            test_cases.extend([reason_test_case, suggestion_test_case, methods_test_case, code_segment_test_case])
    return test_cases

def evaluate_test_cases_with_metrics(test_cases: List[LLMTestCase], metrics: List[GEval], example_id: str) -> List[Dict]:
    results = []

    for test_case in test_cases:
        for metric in metrics:
            metric.measure(test_case)
            results.append({
                "example_id": example_id,
                "test_case": test_case.input,
                "test_case_category": test_case.name,
                "metric": metric.name,
                "score": metric.score,
                "reason": metric.reason
            })
    
    return results

def evaluate_example(example_id: str, client: AzureOpenAI, is_java: bool) -> List[Dict]:
    expected_filepath = f'data/expected/{example_id}_java.json' if is_java else f'data/expected/{example_id}.json'
    response_filepaths = [f'data/responses/{example_id}_{i}java.json' for i in range(5)] if is_java else [f'data/responses/{example_id}_{i}.json' for i in range(5)]

    expected_output = load_review(expected_filepath)
    actual_outputs = load_reviews(response_filepaths)

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

        result_filepath = f'data/results/{example_id}_{i}_java.json' if is_java else f'data/results/{example_id}_{i}.json'
        write_result(result_filepath, results)
        all_results.append(results)
    return all_results 

"""
def evaluate_example(example_id: str, client: AzureOpenAI) -> Dict[str, List[float]]:
    expected_filepath = f'data/expected/{example_id}.json'
    response_filepaths = [f'data/responses/{example_id}_{i}.json' for i in range(5)]
    
    expected_output = load_review(expected_filepath)
    actual_outputs = load_reviews(response_filepaths)

    all_f1_scores = []
    correctness_results = []

    for i, actual_output in enumerate(actual_outputs):
        all_actual_principles = [violation['principle'] for violation in actual_output['principle_violations']]
        all_expected_principles = [violation['principle'] for violation in expected_output['principle_violations']]

        # F1 score test cases
        test_cases_f1 = [LLMTestCase(
                        name=f"F1-Test-Case-{example_id}-{i}",
                        input="Check principle violations in the document.",
                        actual_output=all_actual_principles,
                        expected_output=all_expected_principles
                    )]

        all_metrics = get_all_metrics(client)
        selected_f1_metrics = select_metrics(all_metrics, ["F1-Score"])

        f1_results = evaluate_test_cases_with_metrics(test_cases_f1, selected_f1_metrics, example_id)
        print(f1_results)
        all_f1_scores.extend([result['score'] for result in f1_results if result['metric'] == 'F1-Score'])


        # Correctness test cases
        test_cases_correctness = generate_test_cases(expected_output, actual_output, example_id)
        selected_correctness_metrics = select_metrics(all_metrics, ["Correctness"])

        correctness_results.extend(evaluate_test_cases_with_metrics(test_cases_correctness, selected_correctness_metrics, example_id))

    # Berechnung der Mittelwerte und Standardabweichungen für die F1-Scores
    mean_f1 = np.mean(all_f1_scores)
    std_f1 = np.std(all_f1_scores)

    print(f"{example_id} - Mean F1: {mean_f1}, Std F1: {std_f1}")
    print(all_f1_scores)

    return {
        'f1_scores': {
            'mean': mean_f1,
            'std': std_f1,
            'scores': all_f1_scores
        },
        'correctness_results': correctness_results
    }"""

def evaluate_all_examples(client: AzureOpenAI) -> List[Dict]:
    example_ids = [f"example_{i}_{j}_{k}_java" for i in range(1, 7) for j in range(1, 3) for k in range(0, 5)]
    all_results = []

    for example_id in example_ids:
        print(f"Evaluating example: {example_id}")
        results = evaluate_example(example_id, client)
        all_results.extend(results)

    write_result('data/all_evaluation_results.json', all_results)
    return all_results

def load_all_example_results() -> List[Dict]:
    example_ids_java = [f"example_{i}_{j}_{k}_java" for i in range(1, 7) for j in range(1, 3) for k in range(0, 5)]
    example_ids = [f"example_{i}_{j}_{k}" for i in range(1, 7) for j in range(1, 3) for k in range(5)]
    example_ids.extend(example_ids_java)
    print("ExampleIds: " , example_ids)
    all_results = []

    for example_id in example_ids:
        print(f"Evaluating example: {example_id}")
        results = load_results(example_id)
        all_results.extend(results)

    write_result('data/all_evaluation_results.json', all_results)
    return all_results

def short_label(example_id):
    parts = example_id.split('_')
    return '.'.join(parts[-2:])

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

def plot_f1_scores(f1_scores_data):
    data = {
        "example_id": [],
        "f1_score_mean": [],
        "f1_score_std": []
    }

    for example_id, stats in f1_scores_data.items():
        mean_score = stats['f1']['mean']
        std_score = stats['f1']['std']
        data["example_id"].append(example_id)
        data["f1_score_mean"].append(mean_score)
        data["f1_score_std"].append(std_score)

    df = pd.DataFrame(data)

    plt.figure(figsize=(12, 6))
    sns.barplot(x="example_id", y="f1_score_mean", data=df, capsize=.2, palette="Blues_d")
    plt.errorbar(df["example_id"], df["f1_score_mean"], yerr=df["f1_score_std"], fmt='o', color='r')
    plt.title('Mean F1-Score by Example')
    plt.xlabel('Example ID')
    plt.ylabel('Mean F1-Score')
    plt.xticks(rotation=45)
    plt.show()

    for example_id, stats in f1_scores_data.items():
        print(f"Example ID: {example_id}, Mean F1-Score: {stats['f1']['mean']}, Std Dev: {stats['f1']['std']}")

def aggregate_and_calculate_stats(example_id: str, all_categories: List[str]) -> Dict[str, Dict[str, float]]:
    response_filepaths = [f'data/results/{example_id}_{i}.json' for i in range(5)]
    all_results = load_reviews(response_filepaths)

    f1_scores = []
    correctness_scores = defaultdict(list)

    for results in all_results:
        for result in results:
            test_case_scores = defaultdict(list)
            for i in range(len(result)):
                # Beispiel-ID extrahieren (falls benötigt, kann hier eine Validierung erfolgen)
                example_id = result[i]['example_id']
                
                # F1-Score sammeln
                if result[i]['metric'] == 'F1-Score':
                    f1_scores.append(result[i]['score'])
                
                # Correctness-Metrik analysieren
                if result[i]['metric'] == 'Correctness':
                    test_case = result[i]['test_case']
                    
                    # Kategorie bestimmen
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
            
            # Berechnung von Mean und Std für jeden Testfall
            for category, scores in test_case_scores.items():
                if scores: 
                    mean = np.mean(scores)
                    std = np.std(scores)
                    correctness_scores[category].append({
                        'mean': mean,
                        'std': std,
                        'scores': scores
                    })

    # F1-Statistiken berechnen
    f1_mean = np.mean(f1_scores) if f1_scores else 0
    f1_std = np.std(f1_scores) if f1_scores else 0

    # Correctness-Kategorie-Statistiken berechnen
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

    # Fehlende Kategorien mit Standardwerten auffüllen
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

def short_label(example_id):
    parts = example_id.split('_')
    return '.'.join(parts[-2:])

""" def aggregate_and_calculate_stats(example_id: str) -> Dict[str, Dict[str, float]]:
    response_filepaths = [f'data/results/{example_id}_{i}.json' for i in range(5)]
    all_results = load_reviews(response_filepaths)

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
    for category, stats in correctness_stats.items():
        print(f"Kategorie: {category}")
        print(f"  Mean: {stats['mean']}")
        print(f"  Std: {stats['std']}")
        print(f"  Scores: {stats['scores']}")
        print(f"  Test Cases:")
        for tc in stats['test_cases']:
            print(f"    - Mean: {tc['mean']}, Std: {tc['std']}, Scores: {tc['scores']}") 

    return {
        'f1': {'mean': f1_mean, 'std': f1_std, 'scores': f1_scores},
        'correctness': correctness_stats
    }
 """

""" def aggregate_scores(data: Dict[str, List[float]], all_example_ids: List[str]):
    Helper function to aggregate scores for each example id using the specified method.
    aggregated_data = {key: [] for key in all_example_ids}

    # Populate the aggregated data
    for key, values in data.items():
        aggregated_data[key].extend(values)

    # Calculate mean and standard deviation
    for key in aggregated_data:
        if not aggregated_data[key]:  # Handle empty lists
            aggregated_data[key] = {'mean': 0, 'std_dev': 0}
        else:
            mean_value = np.mean(aggregated_data[key])
            std_dev_value = np.std(aggregated_data[key])  # Use ddof=1 for sample standard deviation
            aggregated_data[key] = {'mean': mean_value, 'std_dev': std_dev_value}
    
    return aggregated_data """

def plot_correctness_scores(correctness_scores_data, all_categories):
    data = {
        "example_id": [],
        "category": [],
        "correctness_mean": [],
        "correctness_std": []
    }

    for example_id, category_stats in correctness_scores_data.items():
        for category in all_categories:
            mean_score = category_stats[category]['mean']
            std_score = category_stats[category]['std']
            data["example_id"].append(example_id)
            data["category"].append(category)
            data["correctness_mean"].append(mean_score)
            data["correctness_std"].append(std_score)

    df = pd.DataFrame(data)

    plt.figure(figsize=(16, 8))
    sns.barplot(x="example_id", y="correctness_mean", hue="category", data=df, dodge=True, capsize=.2, palette="muted")
    for idx, row in df.iterrows():
        plt.errorbar(row["example_id"], row["correctness_mean"], yerr=row["correctness_std"], fmt='o', color='r')
    plt.title('Mean Correctness Score by Example and Category')
    plt.xlabel('Example ID')
    plt.ylabel('Mean Correctness Score')
    plt.xticks(rotation=45)
    plt.legend(title="Category")
    plt.show()

    for example_id, category_stats in correctness_scores_data.items():
        for category in all_categories:
            mean = category_stats[category]['mean']
            std = category_stats[category]['std']
            print(f"Example ID: {example_id}, Category: {category}, Mean Correctness: {mean}, Std Dev: {std}")

def plot_correctness_reasons(results: List[List[Dict]]):
    data_correctness = {
        "example_id": {},
        "example_id_java": {}
    }

    all_example_ids = ['example_1_1', 'example_1_2', 'example_2_1', 'example_2_2', 'example_3_1', 'example_3_2', 'example_4_1', 'example_4_2', 'example_5_1', 'example_5_2', 'example_6_1', 'example_6_2']
    all_example_ids_java = [id_ + '_java' for id_ in all_example_ids]

    for example_results in results:
        for result in example_results:
            if result["metric"] == "Correctness" and "reason" in result["test_case"]:
                if "java" in result["example_id"]:
                    if result["example_id"] not in data_correctness["example_id_java"]:
                        data_correctness["example_id_java"][result["example_id"]] = []
                    data_correctness["example_id_java"][result["example_id"]].append(result["score"])
                else:
                    if result["example_id"] not in data_correctness["example_id"]:
                        data_correctness["example_id"][result["example_id"]] = []
                    data_correctness["example_id"][result["example_id"]].append(result["score"])

    data_correctness["example_id"] = {key: data_correctness["example_id"].get(key, [0]) for key in all_example_ids}
    data_correctness["example_id_java"] = {key: data_correctness["example_id_java"].get(key, [0]) for key in all_example_ids_java}

    aggregated_data_id = aggregate_scores(data_correctness["example_id"], all_example_ids)
    aggregated_data_java = aggregate_scores(data_correctness["example_id_java"], all_example_ids_java)

    correctness_scores_id_sorted = [aggregated_data_id[id_]['mean'] for id_ in all_example_ids]
    correctness_scores_java_sorted = [aggregated_data_java[id_]['mean'] for id_ in all_example_ids_java]

    std = [aggregated_data_id[id_]['std_dev'] for id_ in all_example_ids]
    std_java = [aggregated_data_java[id_]['std_dev'] for id_ in all_example_ids_java]

    print(aggregated_data_id)

    x_labels = [short_label(id_) for id_ in all_example_ids]

    barWidth = 0.3
    r1 = np.arange(len(all_example_ids))
    r2 = [x + barWidth for x in r1]

    plt.figure(figsize=(12, 6))

    plt.bar(r1, correctness_scores_id_sorted, width=barWidth, color='blue', edgecolor='black', yerr=std,label='python')
    plt.bar(r2, correctness_scores_java_sorted, width=barWidth, color='cyan', edgecolor='black', yerr=std_java, label='java')

    plt.xlabel('Examples-ID', fontweight='bold')
    plt.xticks([r + barWidth/2 for r in range(len(all_example_ids))], x_labels, rotation=45)
    plt.ylabel('Correctness Score', fontweight='bold')
    plt.title('Correctness Score of Reason in each example', fontweight='bold')
    plt.legend()

    plt.show()
def plot_correctness_suggestion(results: List[List[Dict]]):
    data_correctness = {
        "example_id": {},
        "example_id_java": {}
    }

    all_example_ids = ['example_1_1', 'example_1_2', 'example_2_1', 'example_2_2', 'example_3_1', 'example_3_2', 'example_4_1', 'example_4_2', 'example_5_1', 'example_5_2', 'example_6_1', 'example_6_2']
    all_example_ids_java = [id_ + '_java' for id_ in all_example_ids]

    for example_results in results:
        for result in example_results:
            if result["metric"] == "Correctness" and "suggestion" in result["test_case"]:
                if "java" in result["example_id"]:
                    if result["example_id"] not in data_correctness["example_id_java"]:
                        data_correctness["example_id_java"][result["example_id"]] = []
                    data_correctness["example_id_java"][result["example_id"]].append(result["score"])
                else:
                    if result["example_id"] not in data_correctness["example_id"]:
                        data_correctness["example_id"][result["example_id"]] = []
                    data_correctness["example_id"][result["example_id"]].append(result["score"])

    data_correctness["example_id"] = {key: data_correctness["example_id"].get(key, [0]) for key in all_example_ids}
    data_correctness["example_id_java"] = {key: data_correctness["example_id_java"].get(key, [0]) for key in all_example_ids_java}

    aggregated_data_id = aggregate_scores(data_correctness["example_id"], all_example_ids)
    aggregated_data_java = aggregate_scores(data_correctness["example_id_java"], all_example_ids_java)

    correctness_scores_id_sorted = [aggregated_data_id[id_] for id_ in all_example_ids]
    correctness_scores_java_sorted = [aggregated_data_java[id_] for id_ in all_example_ids_java]

    x_labels = [short_label(id_) for id_ in all_example_ids]

    barWidth = 0.3
    r1 = np.arange(len(all_example_ids))
    r2 = [x + barWidth for x in r1]

    plt.figure(figsize=(12, 6))

    plt.bar(r1, correctness_scores_id_sorted, width=barWidth, color='blue', edgecolor='black', label='python')
    plt.bar(r2, correctness_scores_java_sorted, width=barWidth, color='cyan', edgecolor='black', label='java')

    plt.xlabel('Examples-ID', fontweight='bold')
    plt.xticks([r + barWidth/2 for r in range(len(all_example_ids))], x_labels, rotation=45)
    plt.ylabel('Correctness Score', fontweight='bold')
    plt.title('Correctness Score of Suggestion in each example', fontweight='bold')
    plt.legend()

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
    client = AzureOpenAI(model=custom_model)
    example_ids = [f"example_{i}_{j}" for i in range(1, 7) for j in range(1, 3)]
    all_categories = ['reason', 'suggestion', 'method names', 'code segments']

    # Evaluate and save results for each example
    for example_id in example_ids:
        print(f"Evaluating example: {example_id}")
        evaluate_example(example_id, client)

    # Aggregating and calculating stats
    aggregated_data = {}
    for example_id in example_ids:
        stats = aggregate_and_calculate_stats(example_id, all_categories)
        aggregated_data[example_id] = stats

    # Plotting F1 results
    plot_f1_scores(aggregated_data)

    # Plotting Correctness results
    # Extract correctness data 
    correctness_data = {example_id: stats['correctness'] for example_id, stats in aggregated_data.items()}
    plot_correctness_scores(correctness_data, all_categories)
    #evaluate_example("example_1_2", azure_openai, False)
    #load_all_example_results(azure_openai)
    #results = load_review("data/all_evaluation_results.json")
    #write_result("all_evaluation_results.json", results)
    #evaluate_all_examples(azure_openai)
    
    #results = load_results("example_1_2")
    #lot_f1_scores(results)
    #plot_correctness_reasons(result)

if __name__ == "__main__":
    main()