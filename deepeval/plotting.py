from matplotlib import pyplot as plt
import numpy as np
import pandas as pd

def short_label(example_id):
    parts = example_id.split('_')
    return '.'.join(parts[-2:])

def plot_f1_scores(f1_scores_data):
    data_python = {
        "example_id": [],
        "f1_score_mean": [],
        "f1_score_std": []
    }
    data_java = {
        "example_id": [],
        "f1_score_mean": [],
        "f1_score_std": []
    }
    
    for example_id, stats in f1_scores_data['python'].items():
        data_python["example_id"].append(example_id)
        data_python["f1_score_mean"].append(stats['f1']['mean'])
        data_python["f1_score_std"].append(stats['f1']['std'])

    for example_id, stats in f1_scores_data['java'].items():
        data_java["example_id"].append(example_id)
        data_java["f1_score_mean"].append(stats['f1']['mean'])
        data_java["f1_score_std"].append(stats['f1']['std'])

    df_python = pd.DataFrame(data_python)
    df_java = pd.DataFrame(data_java)

    x_labels_python = [short_label(id_) for id_ in df_python["example_id"]]
    
    barWidth = 0.3
    r1 = np.arange(len(df_python["example_id"]))
    r2 = [x + barWidth for x in r1]

    plt.figure(figsize=(12, 6))

    plt.bar(r1, df_python["f1_score_mean"], width=barWidth, color='royalblue', edgecolor='black', label='Python', yerr=df_python["f1_score_std"], capsize=3)
    plt.bar(r2, df_java["f1_score_mean"], width=barWidth, color='cyan', edgecolor='black', label='Java', yerr=df_java["f1_score_std"], capsize=3)
    plt.ylim(0, 1.1)
    plt.xlabel('Examples ID', fontweight='bold')
    plt.xticks([r + barWidth/2 for r in range(len(df_python["example_id"]))], x_labels_python)
    plt.ylabel('Mean F1 Score', fontweight='bold')
    plt.title('Mean F1-Score by Example', fontweight='bold')
    plt.legend()

    plt.show()


def plot_correctness(data, category):
    all_example_ids = ['example_1_1', 'example_1_2', 'example_2_1', 'example_2_2', 'example_3_1', 'example_3_2', 'example_4_1', 'example_4_2', 'example_5_1', 'example_5_2', 'example_6_1', 'example_6_2']
    
    data_correctness_python = {
        "example_id": [],
        "correctness_mean": [],
        "correctness_std": []
    }

    data_correctness_java = {
        "example_id": [],
        "correctness_mean": [],
        "correctness_std": []
    }

    for example_id in all_example_ids:
        stats_python = data['python'].get(example_id, {})
        stats_java = data['java'].get(example_id, {})
        
        if stats_python and 'correctness' in stats_python and category in stats_python["correctness"]:
            data_correctness_python["example_id"].append(example_id)
            data_correctness_python["correctness_mean"].append(stats_python["correctness"][category]["mean"])
            data_correctness_python["correctness_std"].append(stats_python["correctness"][category]["std"])
        
        if stats_java and 'correctness' in stats_java and category in stats_java["correctness"]:
            data_correctness_java["example_id"].append(example_id + "_java")
            data_correctness_java["correctness_mean"].append(stats_java["correctness"][category]["mean"])
            data_correctness_java["correctness_std"].append(stats_java["correctness"][category]["std"])

    df_python = pd.DataFrame(data_correctness_python)
    df_java = pd.DataFrame(data_correctness_java)

    x_labels = [short_label(id_) for id_ in df_python["example_id"]]

    barWidth = 0.3
    r1 = np.arange(len(df_python["example_id"]))
    r2 = [x + barWidth for x in r1]

    plt.figure(figsize=(12, 6))

    plt.bar(r1, df_python["correctness_mean"], width=barWidth, color='blue', edgecolor='black', yerr=df_python["correctness_std"], label='Python', capsize=3)
    plt.bar(r2, df_java["correctness_mean"], width=barWidth, color='cyan', edgecolor='black', yerr=df_java["correctness_std"], label='Java', capsize=3)

    plt.xlabel('Examples-ID', fontweight='bold')
    plt.xticks([r + barWidth / 2 for r in range(len(x_labels))], x_labels, rotation=45)
    plt.ylabel('Correctness Score', fontweight='bold')
    plt.title(f'Correctness Score of {category.capitalize()} in each example', fontweight='bold')
    plt.legend()
    plt.ylim(0, 1.1)

    plt.show()


def plot_correctness_scores(data, categories):
    for category in categories:
        plot_correctness(data, category)


def plot_llm_comparison(data):
    models = [item['model'] for item in data]
    python_means = [item['python'][0] for item in data]

    python_conf = [item['python'][1] for item in data]
    # Berechnung der Fehlerbalken (yerr) für Python
    python_lowers = [mean - lower for mean, (lower, upper) in zip(python_means, python_conf)]
    python_uppers = [upper - mean for mean, (lower, upper) in zip(python_means, python_conf)]
    python_yerr = np.array([python_lowers, python_uppers])

    print(python_yerr)

    java_means = [item['java'][0] for item in data]
    java_conf = [item['java'][1] for item in data]
    # Berechnung der Fehlerbalken (yerr) für Java
    java_lowers = [mean - lower for mean, (lower, upper) in zip(java_means, java_conf)]
    java_uppers = [upper - mean for mean, (lower, upper) in zip(java_means, java_conf)]
    java_yerr = np.array([java_lowers, java_uppers])

    barWidth = 0.3
    r1 = np.arange(len(models))
    r2 = [x + barWidth for x in r1]

    

    plt.figure(figsize=(12, 6))
    plt.bar(r1, python_means, width=barWidth, color='royalblue', edgecolor='black', yerr=python_yerr, capsize=3, label='Python')
    plt.bar(r2, java_means, width=barWidth, color='cyan', edgecolor='black', yerr=java_yerr, capsize=3, label='Java')
    plt.ylim(0, 1)
    plt.xlabel('GPT-Models', fontweight='bold')
    plt.xticks([r + barWidth/2 for r in range(len(models))], models)
    plt.ylabel('Mean F1 Score', fontweight='bold')
    plt.title('Mean F1-Score of different GPT-Models', fontweight='bold')
    plt.legend()

    plt.show()
