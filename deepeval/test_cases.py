from typing import Dict, List
from deepeval.test_case import LLMTestCase

def select_test_cases(test_cases: List[LLMTestCase], test_case_names: List[str]) -> List[LLMTestCase]:
    return [test_case for test_case in test_cases if test_case.name in test_case_names]

def generate_test_cases(expected_output: Dict, actual_output: Dict, example_id: str) -> List[LLMTestCase]:
    test_cases = []
    
    expected_violations_dict = {violation['principle']: violation for violation in expected_output['principle_violations']}
    actual_violations_dict = {violation['principle']: violation for violation in actual_output['principle_violations']}

    all_principles = set(expected_violations_dict.keys()).union(set(actual_violations_dict.keys()))
    
    for principle in all_principles:
        expected_violation = expected_violations_dict.get(principle, {})
        actual_violation = actual_violations_dict.get(principle, {})
        
        # Check if principle exists in both expected and actual
        if principle in expected_violations_dict and principle in actual_violations_dict:
            # Create test cases for reason, suggestion, method names
            reason_test_case = LLMTestCase(
                name=f"Reason-Test-Case-{principle}",
                input=f"Review reason for {principle} in {example_id}",
                actual_output=actual_violation.get('reason', ''),
                expected_output=expected_violation.get('reason', '')
            )
            suggestion_test_case = LLMTestCase(
                name=f"Suggestion-Test-Case-{principle}",
                input=f"Review suggestion for {principle} in {example_id}",
                actual_output=actual_violation.get('suggestion', ''),
                expected_output=expected_violation.get('suggestion', '')
            )
            methods_test_case = LLMTestCase(
                name=f"Method-Name-Test-Case-{principle}",
                input=f"Review method names for {principle} in {example_id}",
                actual_output=actual_violation.get('method_names', []),
                expected_output=expected_violation.get('method_names', [])
            )
            code_segment_test_case = LLMTestCase(
                name=f"Changes-Test-Case-{principle}",
                input=f"Review code segments for {principle} in {example_id}",
                actual_output=actual_violation.get('changes', []),
                expected_output=expected_violation.get('changes', [])
            )
            
            test_cases.extend([reason_test_case, suggestion_test_case, methods_test_case, code_segment_test_case])
    return test_cases