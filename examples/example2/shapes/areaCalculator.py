import math
import json

class AreaCalculator:
    def __init__(self, shapes=[]) -> None:
        self.shapes = shapes
    
    def sum(self):
        areas = []
        for shape in self.shapes:
            if isinstance(shape, Square):
                areas.append(math.pow(shape.length, 2))
            elif isinstance(shape, Circle):
                areas.append(math.pi * math.pow(shape.radius, 2))
        return sum(areas)

class Square:
    def __init__(self, length) -> None:
        self.length = length

class Circle:
    def __init__(self, radius) -> None:
        self.radius = radius

class SumCalculator:
    def __init__(self, areaCalculator: AreaCalculator) -> None:
        self.calculator = areaCalculator

    def JSON(self):
        data = {
            'sum': self.calculator.sum()
        }
        return json.dumps(data)

    def HTML(self):
        return f"Sum of the areas of provided shapes: {self.calculator.sum()}"


# Beispiel für die Verwendung:
shapes = [
    Circle(2),
    Square(5),
    Square(6)
]

areas = AreaCalculator(shapes)
output = SumCalculator(areas)

# Ausgabe der Ergebnisse
print(output.JSON())
print(output.HTML())