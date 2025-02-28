
from common_config import *
from data_loader import load_json_files, write_result


def calculate_total_tokens_by_example(example_id: str, is_java: bool, file_dir: str) -> int:
    response_filepaths = [f'{file_dir}/{example_id}_{i}_java.json' for i in range(COUNTER)] if is_java else [f'{file_dir}/{example_id}_{i}.json' for i in range(COUNTER)]
    all_results = load_json_files(response_filepaths)

    total_tokens = 0
    for results in all_results:
        if file_dir == RESPONSES_DATA_DIR:
            total_tokens += results['usage']['total_tokens']
        elif file_dir == RESULTS_DATA_DIR:
            for result in results:
                for item in result:
                    total_tokens += item['tokens_used']
            
    return total_tokens

def calculate_total_tokens():
    total_tokens = 0
    total_tokens_java = 0
    total_tokens_python = 0
    result_tokens_python = 0
    response_tokens_python = 0
    result_tokens_java = 0
    response_tokens_java = 0


    example_ids = [f"example_{i}_{j}" for i in range(1, 7) for j in range(1, 3)]
    for example_id in example_ids:
        result_tokens_python += calculate_total_tokens_by_example(example_id, False, RESULTS_DATA_DIR)
        response_tokens_python += calculate_total_tokens_by_example(example_id, False, RESPONSES_DATA_DIR)

        result_tokens_java += calculate_total_tokens_by_example(example_id, True, RESULTS_DATA_DIR)
        response_tokens_java += calculate_total_tokens_by_example(example_id, True, RESPONSES_DATA_DIR)

    total_tokens_python = result_tokens_python + response_tokens_python
    total_tokens_java = result_tokens_java + response_tokens_java
    total_tokens = total_tokens_python + total_tokens_java
    print("Total Tokens: ", total_tokens)
    result = {
        "response_tokens_python": response_tokens_python,
        "response_tokens_java": response_tokens_java,
        "result_tokens_python": result_tokens_python,
        "result_tokens_java": result_tokens_java,
        "total_tokens_python": total_tokens_python,
        "total_tokens_java": total_tokens_java,
        "total_tokens": total_tokens
    }
    write_result(f'{RESPONSES_DATA_DIR}/total_tokens.json', result)
    write_result(f'{RESULTS_DATA_DIR}/total_tokens.json', result)
    return total_tokens