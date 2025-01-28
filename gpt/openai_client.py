import json
from config import API_KEY, API_BASE, DEPLOYMENT_NAME, API_VERSION
from openai import AzureOpenAI
from models import CodeReview
import system_prompt

class AzureOpenAIClient:
    def __init__(self):
        self.client = AzureOpenAI(azure_endpoint=API_BASE, api_key=API_KEY, api_version=API_VERSION)
    
    def request_code_review(self, user_prompt: str):
        response = self.client.beta.chat.completions.parse(
            model=DEPLOYMENT_NAME,
            messages=[
                {
                    "role": "system",
                    "content": system_prompt.SYSTEM_PROMPT
                },
                {
                    "role": "user",
                    "content": user_prompt
                }
            ],
            response_format=CodeReview)
        return response
    
    def save_code_review_response(self, filename: str, completion):
        message = completion.choices[0].message.parsed
        usage = completion.usage.dict(include={"completion_tokens": True, "prompt_tokens": True, "total_tokens": True})
        combined_response = {**message.dict(), "usage": usage}

        with open(filename, 'w') as file:
            json.dump(combined_response, file, indent=4)
    
    def print_code_review_output(self, completion) -> None:
        message = completion.choices[0].message
        if message.parsed:
            print('-' * 40)
            print("Feedback: " + message.parsed.overall_feedback)
            print('-' * 40)
            for violation in message.parsed.principle_violations:
                for key, value in violation:
                    if key == 'code':
                        print(f"{key.capitalize()}:")
                        for segment in value:
                            print(segment)
                            print('-' * 20)
                    else:
                        print(f"{key.capitalize()}: {value}")
                    
                print('-' * 40)
               # Ausgabe der verwendeten Tokens
            print("Usage Information:")
            print(f"Prompt Tokens: {completion.usage.prompt_tokens}")
            print(f"Completion Tokens: {completion.usage.completion_tokens}")
            print(f"Total Tokens: {completion.usage.total_tokens}")
        else:
            print(message.refusal)