import os
from dotenv import load_dotenv

project_root = os.path.abspath(os.path.join(os.path.dirname(__file__)))
dotenv_path = os.path.join(project_root, '.env')
load_dotenv(dotenv_path)

API_BASE = os.getenv("AZURE_OPENAI_API_BASE")
API_KEY = os.getenv("AZURE_OPENAI_API_KEY")
API_VERSION = os.getenv("AZURE_OPENAI_API_VERSION")
DEPLOYMENT_NAME = os.getenv("AZURE_OPENAI_DEPLOYMENT_NAME")

# Directory paths
EXPECTED_DATA_DIR = 'data/expected'
RESPONSES_DATA_DIR = 'data/responses'
RESULTS_DATA_DIR = 'data/results'

# Other configurations
COUNTER = 5

print(API_VERSION, DEPLOYMENT_NAME)