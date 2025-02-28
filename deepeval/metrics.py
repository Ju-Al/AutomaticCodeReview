from typing import List
from client import AzureOpenAI
from deepeval.metrics import GEval
from deepeval.test_case import LLMTestCaseParams

def select_metrics(all_metrics: List[GEval], metrics_to_select: List[str]) -> List[GEval]:
    return [metric for metric in all_metrics if metric.name in metrics_to_select]

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