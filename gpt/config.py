import os
from dotenv import load_dotenv

load_dotenv()

API_BASE = os.getenv("AZURE_OPENAI_API_BASE")
DEPLOYMENT_NAME = os.getenv("AZURE_OPENAI_DEPLOYMENT_NAME")
API_KEY = os.getenv("AZURE_OPENAI_API_KEY")
API_VERSION = os.getenv("AZURE_OPENAI_API_VERSION")

if not all([API_BASE, DEPLOYMENT_NAME, API_KEY, API_VERSION]):
    raise EnvironmentError("Please set the environment variables in the .env file")