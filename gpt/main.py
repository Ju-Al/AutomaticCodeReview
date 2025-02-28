
from glob import glob
import os
import time
from openai_client import AzureOpenAIClient
from common_config import RESPONSES_DATA_DIR, COUNTER

def get_user_prompts(file_paths):
    contents = []
    for file_path in file_paths:
        with open(file_path, 'r') as file:
            contents.append(file.read())
    return "\n".join(contents)

def process_files(file_paths, example_id):
    client = AzureOpenAIClient()
    user_prompt = get_user_prompts(file_paths)
    start_time = time.time()
    completion = client.request_code_review(user_prompt)
    end_time = time.time()
    execution_time = end_time - start_time

    client.print_code_review_output(completion)
    client.save_code_review_response(f'{RESPONSES_DATA_DIR}/{example_id}.json', completion, execution_time)

def collect_file_paths(directory, ext):
    file_paths = []
    file_paths.extend(glob(os.path.join(directory, '**', f'*.{ext}'), recursive=True))
    return file_paths

def main():

    """ base_dirs = [
            ('examples/example1/order', 'example_1_1'),
            ('examples/example1/shapes', 'example_1_2'),
            ('examples/example2/order', 'example_2_1'),
            ('examples/example2/shapes', 'example_2_2'),
            ('examples/example3/order', 'example_3_1'),
            ('examples/example3/shapes', 'example_3_2'),
            ('examples/example4/order', 'example_4_1'),
            ('examples/example4/shapes', 'example_4_2'),
            ('examples/example5/order', 'example_5_1'),
            ('examples/example5/user', 'example_5_2'),
            ('examples/example6/invoiceCalculator', 'example_6_1'),
            ('examples/example6/fitnesse', 'example_6_2'),
        ] """
    base_dirs = [
            ('examples/example4/shapes', 'example_4_2')]
    
    for directory, example_id in base_dirs:
        for i in range(COUNTER):
            if not os.path.exists(f'{RESPONSES_DATA_DIR}/{example_id}_{i}.json'):
                file_paths = collect_file_paths(directory, 'py')
                print(f"Processing files for {example_id}_{i} from {directory} with files {file_paths}")
                process_files(file_paths, f'{example_id}_{i}')
            if not os.path.exists(f'{RESPONSES_DATA_DIR}/{example_id}_{i}_java.json'):
                file_paths = collect_file_paths(directory, 'java')
                print(f"Processing files for {example_id}_{i} from {directory} with files {file_paths}")
                process_files(file_paths, f'{example_id}_{i}_java')
            else:
                print(f"file for {example_id}_{i} already exist")

if __name__ == "__main__":
    main()