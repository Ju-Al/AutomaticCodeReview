import os
from dotenv import load_dotenv
from langchain_openai import AzureChatOpenAI
from deepeval.models.base_model import DeepEvalBaseLLM

class AzureOpenAI(DeepEvalBaseLLM):
    def __init__(
        self,
        model
    ):
        self.model = model

    def load_model(self):
        return self.model

    def generate(self, prompt: str) -> str:
        chat_model = self.load_model()
        return chat_model.invoke(prompt).content

    async def a_generate(self, prompt: str) -> str:
        chat_model = self.load_model()
        res = await chat_model.ainvoke(prompt)
        return res.content

    def get_model_name(self):
        return "Custom Azure OpenAI Model"

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