from enum import Enum
from typing import List
from pydantic import BaseModel

class Principle(str, Enum):
    SRP = "Single Responsibility Principle"
    OCP = "Open/Closed Principle"
    LSP = "Liskov Substitution Principle"
    ISP = "Interface Segregation Principle"
    DIP = "Dependency Inversion Principle"
    DRY = "Don't Repeat Yourself"

class Changes(BaseModel):
    original_code: List[str]
    suggestion_code: List[str]

class PrincipleViolation(BaseModel):
    principle: Principle
    lines: List[List [int]]
    method_names: List[str]
    reason: str
    suggestion: str
    changes: List[Changes]

class CodeReview(BaseModel):
    principle_violations: List[PrincipleViolation]
    overall_feedback: str