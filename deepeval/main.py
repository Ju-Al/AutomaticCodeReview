
from langchain_openai import AzureChatOpenAI
import numpy as np

from client import AzureOpenAI
from common_config import *
from plotting import plot_correctness_scores, plot_f1_scores, plot_llm_comparison
from tokens import calculate_total_tokens
from evaluation import evaluate_example
from stats_calulation import aggregate_and_calculate_stats
import scipy.stats as st 

def mean_confidence_interval(data, confidence=0.95):
    a = 1.0 * np.array(data)
    n = len(a)
    m, se = np.mean(a), st.sem(a)
    h = se * st.t.ppf((1 + confidence) / 2., n-1)
    return m, m-h, m+h

def main():
    custom_model = AzureChatOpenAI(
        openai_api_version=API_VERSION,
        azure_deployment=DEPLOYMENT_NAME,
        azure_endpoint=API_BASE,
        openai_api_key=API_KEY,
    )
    client = AzureOpenAI(model=custom_model)
    example_ids = [f"example_{i}_{j}" for i in range(1, 7) for j in range(1, 3)]
    all_categories = ['reason', 'suggestion', 'method names', 'code segments']

    """ for example_id in example_ids:
        for i in range(COUNTER):
            if not os.path.exists(f'{RESULTS_DATA_DIR}/{example_id}_{i}.json'):
                print(f"Evaluating example: {example_id}_{i}")
                evaluate_example(example_id, client, False)
            if not os.path.exists(f'{RESULTS_DATA_DIR}/{example_id}_{i}_java.json'):
                print(f"Evaluating example: {example_id}_{i}_java")
                evaluate_example(example_id, client, True)
            else:
                print(f"file for {example_id} already exist") """
    # Aggregating and calculating stats
    aggregated_data = {
        'model': {},
        'python': {},
        'java': {}
    }
    models = ['gpt_4o_08_06', 'gpt_4o_11_20', 'gpt_4o_mini_07_18', 'o1_12_17', 'o1_mini_09_12']
    result = []

    
    for model in models:
        python_scores = []
        java_scores = []

        for example_id in example_ids:
            stats = aggregate_and_calculate_stats(example_id, all_categories, False, model)
            stats_java = aggregate_and_calculate_stats(example_id, all_categories, True, model)

            python_scores.extend(stats['f1']['scores'])
            java_scores.extend(stats_java['f1']['scores'])

        python_mean = np.mean(python_scores) if python_scores else 0
        java_mean = np.mean(java_scores) if java_scores else 0

        n_python = len(1.0 * np.array(python_scores))
        n_java = len(java_scores)

        python_confidence = st.t.interval(confidence=0.95, df=n_python-1, loc=python_mean, scale=st.sem(python_scores))
        java_confidence = st.t.interval(confidence=0.95, df=n_java-1, loc=java_mean, scale=st.sem(python_scores)) 

        aggregated_data = {
            'model': model,
            'python': [python_mean, python_confidence],
            'java': [java_mean, java_confidence]
        }
        result.append(aggregated_data)

    # Plotting F1 results
    #plot_llm_comparison(result)
    #plot_correctness_scores(aggregated_data, all_categories)
    #calculate_total_tokens()
    evaluate_example("example_5_1", client, False)


if __name__ == "__main__":
    main()