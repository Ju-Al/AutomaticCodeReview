
from openai_client import AzureOpenAIClient

def get_user_prompts(file_paths):
    contents = []
    for file_path in file_paths:
        with open(file_path, 'r') as file:
            contents.append(file.read())
    return "\n".join(contents)

def process_files(file_paths):
    client = AzureOpenAIClient()
    user_prompt = get_user_prompts(file_paths)
    completion = client.request_code_review(user_prompt)
    client.print_code_review_output(completion)
    client.save_code_review_response("data/responses/example_1.json", completion)

def main():
    input_files = [
        'example1/person/ValidatePerson.java',
        'example1/person/Main.java'
    ]
    process_files(input_files)

if __name__ == "__main__":
    main()